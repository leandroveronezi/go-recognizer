package recognizer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"sync"

	goFace "github.com/leandroveronezi/go-face"
)

// Data descriptor of the human face.
type Data struct {
	Id         string
	Descriptor goFace.Descriptor
}

// Face holds coordinates and descriptor of the human face.
type Face struct {
	Data
	Rectangle image.Rectangle
	// Shapes holds the facial landmark points found by the shape predictor
	// model (5 points with the default shape_predictor_5_face_landmarks.dat).
	Shapes []image.Point
	// Distance is the squared Euclidean distance between this face's
	// descriptor and the matched Dataset entry's descriptor. Only set by
	// Classify/ClassifyMultiples; zero otherwise.
	Distance float64
	// Confidence is a convenience score in [0,1], normalized as
	// 1-Distance/Tolerance -- not a calibrated probability. Only set by
	// Classify/ClassifyMultiples; zero otherwise.
	Confidence float64
}

/*
ModelFiles selects non-default model file names, all resolved relative
to the modelDir passed to Init. Set fields before calling Init; they
have no effect afterward, since model loading happens inside Init.
*/
type ModelFiles struct {
	// Landmark is the shape predictor file name. Empty (the default)
	// uses go-face's shape_predictor_5_face_landmarks.dat, which returns
	// 5 landmark points (eye corners and nose base) in Face.Shapes. Set
	// to "shape_predictor_68_face_landmarks.dat" for full facial contour
	// landmarks (jawline, eyebrows, nose bridge, eyes, lips) instead --
	// it's a much larger download and slightly slower per face, so only
	// opt in if you need the extra landmark detail.
	Landmark string
	// Descriptor is the face-descriptor (ResNet) model file name. Empty
	// uses go-face's default dlib_face_recognition_resnet_model_v1.dat.
	Descriptor string
	// CNN is the CNN face detector model file name, used when UseCNN is
	// true. Empty uses go-face's default mmod_human_face_detector.dat.
	CNN string
}

/*
A Recognizer creates face descriptors for provided images and
classifies them into categories.
*/
type Recognizer struct {
	Tolerance float32
	rec       *goFace.Recognizer
	UseCNN    bool
	UseGray   bool
	// Model selects non-default model file names. Set before Init.
	Model ModelFiles
	// Dataset holds the known face samples. Mutate it only through
	// AddImageToDataset (keeps the classifier in sync automatically) or
	// LoadDataset (call SetSamples afterward -- see SetSamples).
	Dataset []Data
	mu      sync.RWMutex
}

/*
Init initialise a recognizer interface. Set Model before calling Init to
choose non-default model files (e.g. Model.Landmark for full facial
contour landmarks instead of the default 5 points).
*/
func (_this *Recognizer) Init(Path string) error {

	_this.Tolerance = 0.4
	_this.UseCNN = false
	_this.UseGray = true

	_this.mu.Lock()
	_this.Dataset = make([]Data, 0)
	_this.mu.Unlock()

	rec, err := goFace.NewRecognizerWithModels(Path, _this.Model.Landmark, _this.Model.Descriptor, _this.Model.CNN)

	if err == nil {
		_this.rec = rec
	}

	return err

}

/*
Close frees resources taken by the Recognizer. Safe to call multiple
times. Don't use Recognizer after close call.
*/
func (_this *Recognizer) Close() {

	_this.rec.Close()

}

/*
detect runs face detection on Path, returning every face found. It
prefers the raw-pixel path when UseGray is set -- decoding, grayscaling,
and handing pixels straight to goFace, with no JPEG round trip involved
-- and otherwise hands Path straight to go-face, which only understands
JPEG.
*/
func (_this *Recognizer) detect(Path string) ([]goFace.Face, error) {

	if _this.UseGray {

		pixels, width, height, err := _this.loadPixels(Path)
		if err != nil {
			return nil, err
		}

		if _this.UseCNN {
			return _this.rec.RecognizeRawCNN(pixels, width, height)
		}
		return _this.rec.RecognizeRaw(pixels, width, height)

	}

	if _this.UseCNN {
		return _this.rec.RecognizeFileCNN(Path)
	}
	return _this.rec.RecognizeFile(Path)

}

/*
AddImageToDataset add a sample image to the dataset.

The new entry is appended to the underlying classifier immediately (via
goFace.AppendSample), so it's classifiable right away -- no need to call
SetSamples afterward. SetSamples is still required after LoadDataset or
after mutating Dataset directly, since those don't go through this
incremental path.
*/
func (_this *Recognizer) AddImageToDataset(Path string, Id string) error {

	faces, err := _this.detect(Path)

	if err != nil {
		return err
	}

	if len(faces) == 0 {
		return errors.New("Not a face on the image")
	}

	if len(faces) > 1 {
		return errors.New("Not a single face on the image")
	}

	f := Data{}
	f.Id = Id
	f.Descriptor = faces[0].Descriptor

	_this.mu.Lock()
	_this.Dataset = append(_this.Dataset, f)
	idx := len(_this.Dataset) - 1
	_this.mu.Unlock()

	_this.rec.AppendSample(f.Descriptor, int32(idx))

	return nil

}

/*
SetSamples rebuilds the classifier's sample set from the entire current
Dataset.

AddImageToDataset already keeps the classifier in sync incrementally, so
this is only needed after LoadDataset (bulk load) or after mutating
Dataset directly -- those don't go through AddImageToDataset's
incremental path, so the classifier would otherwise keep matching
against a stale or empty sample set.
*/
func (_this *Recognizer) SetSamples() {

	var samples []goFace.Descriptor
	var avengers []int32

	_this.mu.RLock()
	for i, f := range _this.Dataset {
		samples = append(samples, f.Descriptor)
		avengers = append(avengers, int32(i))
	}
	_this.mu.RUnlock()

	_this.rec.SetSamples(samples, avengers)

}

/*
RecognizeSingle returns face if it's the only face on the image or nil otherwise.
*/
func (_this *Recognizer) RecognizeSingle(Path string) (goFace.Face, error) {

	var idFace *goFace.Face
	var err error

	if _this.UseGray {

		pixels, width, height, lerr := _this.loadPixels(Path)
		if lerr != nil {
			return goFace.Face{}, lerr
		}

		if _this.UseCNN {
			idFace, err = _this.rec.RecognizeSingleRawCNN(pixels, width, height)
		} else {
			idFace, err = _this.rec.RecognizeSingleRaw(pixels, width, height)
		}

	} else {

		if _this.UseCNN {
			idFace, err = _this.rec.RecognizeSingleFileCNN(Path)
		} else {
			idFace, err = _this.rec.RecognizeSingleFile(Path)
		}

	}

	if err != nil {
		return goFace.Face{}, fmt.Errorf("Can't recognize: %v", err)

	}
	if idFace == nil {
		return goFace.Face{}, fmt.Errorf("Not a single face on the image")
	}

	return *idFace, nil

}

/*
RecognizeMultiples returns all faces found on the provided image, sorted from
left to right. Empty list is returned if there are no faces, error is
returned if there was some error while decoding/processing image.
*/
func (_this *Recognizer) RecognizeMultiples(Path string) ([]goFace.Face, error) {

	idFaces, err := _this.detect(Path)

	if err != nil {
		return nil, fmt.Errorf("Can't recognize: %v", err)
	}

	return idFaces, nil

}

/*
Classify returns all faces identified in the image. Empty list is returned if no match.

Matches against the sample set from the most recent SetSamples call, not
necessarily the current Dataset -- see SetSamples.
*/
func (_this *Recognizer) Classify(Path string) ([]Face, error) {

	face, err := _this.RecognizeSingle(Path)

	if err != nil {
		return nil, err
	}

	personID := _this.rec.ClassifyThreshold(face.Descriptor, _this.Tolerance)
	if personID < 0 {
		return nil, fmt.Errorf("Can't classify")
	}

	_this.mu.RLock()
	defer _this.mu.RUnlock()

	if personID >= len(_this.Dataset) {
		return nil, fmt.Errorf("Can't classify")
	}

	matched := _this.Dataset[personID]
	distance := goFace.SquaredEuclideanDistance(matched.Descriptor, face.Descriptor)

	facesRec := make([]Face, 0)
	aux := Face{
		Data:       matched,
		Rectangle:  face.Rectangle,
		Shapes:     face.Shapes,
		Distance:   distance,
		Confidence: confidence(distance, _this.Tolerance),
	}
	facesRec = append(facesRec, aux)

	return facesRec, nil

}

/*
ClassifyMultiples returns all faces identified in the image. Empty list is returned if no match.

Matches against the sample set from the most recent SetSamples call, not
necessarily the current Dataset -- see SetSamples.
*/
func (_this *Recognizer) ClassifyMultiples(Path string) ([]Face, error) {

	faces, err := _this.RecognizeMultiples(Path)

	if err != nil {
		return nil, fmt.Errorf("Can't recognize: %v", err)
	}

	facesRec := make([]Face, 0)

	for _, f := range faces {

		personID := _this.rec.ClassifyThreshold(f.Descriptor, _this.Tolerance)
		if personID < 0 {
			continue
		}

		_this.mu.RLock()
		if personID >= len(_this.Dataset) {
			_this.mu.RUnlock()
			continue
		}
		matched := _this.Dataset[personID]
		distance := goFace.SquaredEuclideanDistance(matched.Descriptor, f.Descriptor)
		aux := Face{
			Data:       matched,
			Rectangle:  f.Rectangle,
			Shapes:     f.Shapes,
			Distance:   distance,
			Confidence: confidence(distance, _this.Tolerance),
		}
		_this.mu.RUnlock()

		facesRec = append(facesRec, aux)

	}

	return facesRec, nil

}

/*
confidence normalizes a squared-Euclidean match distance against the
tolerance used to accept it, into a convenience [0,1] score where 0
distance is 1.0 and a distance at (or past) the tolerance cutoff is 0.0.
This is not a calibrated probability.
*/
func confidence(distance float64, tolerance float32) float64 {

	if tolerance <= 0 {
		return 0
	}

	c := 1 - distance/float64(tolerance)

	if c < 0 {
		return 0
	}

	if c > 1 {
		return 1
	}

	return c

}

/*
fileExists check se file exist
*/
func fileExists(FileName string) bool {
	file, err := os.Stat(FileName)
	return (err == nil) && !file.IsDir()
}

/*
jsonMarshal Marshal interface to array of byte
*/
func jsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
