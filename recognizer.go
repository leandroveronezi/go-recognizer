package recognizer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Kagami/go-face"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
)

type Data struct {
	Id         string
	Descriptor face.Descriptor
}

type Face struct {
	Data
	Rectangle image.Rectangle
}

type Recognizer struct {
	Tolerance float32
	rec       *face.Recognizer
	UseCNN    bool
	UseGray   bool
	Dataset   []Data
}

func (_this *Recognizer) SaveDataset(Path string) error {

	data, err := jsonMarshal(_this.Dataset)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(Path, data, 0777)

}

func (_this *Recognizer) LoadDataset(Path string) error {

	if !fileExists(Path) {
		return errors.New("file not found")
	}

	file, err := os.OpenFile(Path, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}

	Dataset := make([]Data, 0)
	err = json.NewDecoder(file).Decode(&Dataset)
	if err != nil {
		return err
	}

	_this.Dataset = append(_this.Dataset, Dataset...)

	return nil

}

func (_this *Recognizer) Init(Path string) error {

	_this.Tolerance = 0.4
	_this.UseCNN = false
	_this.UseGray = true

	_this.Dataset = make([]Data, 0)

	rec, err := face.NewRecognizer(Path)

	if err == nil {
		_this.rec = rec
	}

	return err

}

func (_this *Recognizer) Close() {

	_this.rec.Close()

}

func (_this *Recognizer) AddFile(Path string, Id string) error {

	file := Path
	var err error

	if _this.UseGray {

		err, file = _this.createTempGrayFile(file, Id)

		if err != nil {
			return err
		}

		defer os.Remove(file)

	}

	faces, err := _this.rec.RecognizeFileCNN(file)
	if err != nil {
		return err
	}

	if len(faces) > 1 {
		return errors.New("Not a single face on the image")
	}

	f := Data{}
	f.Id = Id
	f.Descriptor = faces[0].Descriptor

	_this.Dataset = append(_this.Dataset, f)

	return nil

}

func (_this *Recognizer) LoadSamples() {

	var samples []face.Descriptor
	var avengers []int32

	for i, f := range _this.Dataset {
		samples = append(samples, f.Descriptor)
		avengers = append(avengers, int32(i))
	}

	_this.rec.SetSamples(samples, avengers)

}

func (_this *Recognizer) Classify(Path string) (error, Data, []Face) {

	file := Path
	var err error

	if _this.UseGray {

		err, file = _this.createTempGrayFile(file, "64ab59ac42d69274f06eadb11348969e")

		if err != nil {
			return err, Data{}, nil
		}

		defer os.Remove(file)

	}

	face, err := _this.rec.RecognizeSingleFileCNN(file)
	if err != nil {
		return fmt.Errorf("Can't recognize: %v", err), Data{}, nil

	}
	if face == nil {
		return fmt.Errorf("Not a single face on the image"), Data{}, nil
	}

	personID := _this.rec.ClassifyThreshold(face.Descriptor, _this.Tolerance)
	if personID < 0 {
		return fmt.Errorf("Can't classify"), Data{}, nil
	}

	facesRec := make([]Face, 0)
	aux := Face{Data: _this.Dataset[personID], Rectangle: face.Rectangle}
	facesRec = append(facesRec, aux)

	return nil, _this.Dataset[personID], facesRec

}

func (_this *Recognizer) ClassifyMultiples(Path string) (error, []Data, []Face) {

	file := Path
	var err error

	if _this.UseGray {

		err, file = _this.createTempGrayFile(file, "64ab59ac42d69274f06eadb11348969e")

		if err != nil {
			return err, nil, nil
		}

		defer os.Remove(file)

	}

	faces, err := _this.rec.RecognizeFileCNN(file)
	if err != nil {
		return fmt.Errorf("Can't recognize: %v", err), nil, nil
	}

	people := make([]Data, 0)

	facesRec := make([]Face, 0)

	for _, f := range faces {

		personID := _this.rec.ClassifyThreshold(f.Descriptor, _this.Tolerance)
		if personID < 0 {
			continue
		}

		people = append(people, _this.Dataset[personID])

		aux := Face{Data: _this.Dataset[personID], Rectangle: f.Rectangle}

		facesRec = append(facesRec, aux)

	}

	return nil, people, facesRec

}

func fileExists(FileName string) bool {
	file, err := os.Stat(FileName)
	return (err == nil) && !file.IsDir()
}

func jsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
