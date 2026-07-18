package recognizer

import (
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const (
	testModelsDir = "examples/models"
	testFotosDir  = "examples/fotos"
)

func newTestRecognizer(t *testing.T) *Recognizer {
	t.Helper()

	rec := &Recognizer{}
	if err := rec.Init(testModelsDir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(rec.Close)

	return rec
}

func TestInitDefaultLandmarks(t *testing.T) {
	rec := newTestRecognizer(t)

	f, err := rec.RecognizeSingle(filepath.Join(testFotosDir, "amy.jpg"))
	if err != nil {
		t.Fatalf("RecognizeSingle: %v", err)
	}

	if len(f.Shapes) != 5 {
		t.Errorf("got %d landmark points, want 5 (default shape predictor)", len(f.Shapes))
	}
}

func TestInitWithShapePredictor68Points(t *testing.T) {
	path := filepath.Join(testModelsDir, "shape_predictor_68_face_landmarks.dat")
	if _, err := os.Stat(path); err != nil {
		t.Skip("shape_predictor_68_face_landmarks.dat not present (optional, ~95MB, not checked into the repo)")
	}

	rec := &Recognizer{}
	rec.Model.Landmark = "shape_predictor_68_face_landmarks.dat"
	if err := rec.Init(testModelsDir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(rec.Close)

	f, err := rec.RecognizeSingle(filepath.Join(testFotosDir, "amy.jpg"))
	if err != nil {
		t.Fatalf("RecognizeSingle: %v", err)
	}

	if len(f.Shapes) != 68 {
		t.Errorf("got %d landmark points, want 68", len(f.Shapes))
	}
}

// TestModelDescriptorFileIsHonored proves Model.Descriptor actually
// controls which file gets loaded, rather than being silently ignored: a
// wrong file name must fail to load, and a genuinely renamed resnet model
// must load successfully and still produce working descriptors.
func TestModelDescriptorFileIsHonored(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"shape_predictor_5_face_landmarks.dat", "mmod_human_face_detector.dat"} {
		copyFile(t, filepath.Join(testModelsDir, name), filepath.Join(dir, name))
	}
	copyFile(t, filepath.Join(testModelsDir, "dlib_face_recognition_resnet_model_v1.dat"), filepath.Join(dir, "my_resnet.dat"))

	bad := &Recognizer{}
	bad.Model.Descriptor = "does-not-exist.dat"
	if err := bad.Init(dir); err == nil {
		t.Fatalf("Init: expected an error for a nonexistent Model.Descriptor file, got nil")
	}

	rec := &Recognizer{}
	rec.Model.Descriptor = "my_resnet.dat"
	if err := rec.Init(dir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(rec.Close)

	f, err := rec.RecognizeSingle(filepath.Join(testFotosDir, "amy.jpg"))
	if err != nil {
		t.Fatalf("RecognizeSingle: %v", err)
	}

	var zero [128]float32
	if f.Descriptor == zero {
		t.Errorf("Descriptor is all-zero, want a real face descriptor")
	}
}

// TestModelCNNFileIsHonored is the same proof as
// TestModelDescriptorFileIsHonored, for Model.CNN.
func TestModelCNNFileIsHonored(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"shape_predictor_5_face_landmarks.dat", "dlib_face_recognition_resnet_model_v1.dat"} {
		copyFile(t, filepath.Join(testModelsDir, name), filepath.Join(dir, name))
	}
	copyFile(t, filepath.Join(testModelsDir, "mmod_human_face_detector.dat"), filepath.Join(dir, "my_cnn.dat"))

	bad := &Recognizer{}
	bad.Model.CNN = "does-not-exist.dat"
	if err := bad.Init(dir); err == nil {
		t.Fatalf("Init: expected an error for a nonexistent Model.CNN file, got nil")
	}

	rec := &Recognizer{}
	rec.Model.CNN = "my_cnn.dat"
	if err := rec.Init(dir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(rec.Close)

	rec.UseCNN = true
	faces, err := rec.RecognizeMultiples(filepath.Join(testFotosDir, "amy.jpg"))
	if err != nil {
		t.Fatalf("RecognizeMultiples (CNN): %v", err)
	}
	if len(faces) == 0 {
		t.Errorf("got 0 faces via the renamed CNN model, want at least 1")
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0600); err != nil {
		t.Fatalf("WriteFile(%s): %v", dst, err)
	}
}

// TestAddImageToDatasetAcceptsPNGWhenUseGray is a regression test for the
// upstream-reported limitation in Kagami/go-face#45 ("Cannot recognize PNG
// files"): go-face's own file loader only understands JPEG, but with the
// default UseGray=true, AddImageToDataset decodes the source with Go's
// standard image package (which supports PNG here via fogleman/gg's
// transitive import of image/png) and re-encodes it as JPEG before handing
// it to go-face, so a PNG source works without extra steps.
func TestAddImageToDatasetAcceptsPNGWhenUseGray(t *testing.T) {
	pngPath := filepath.Join(t.TempDir(), "amy.png")
	convertJPEGToPNG(t, filepath.Join(testFotosDir, "amy.jpg"), pngPath)

	rec := newTestRecognizer(t)
	if err := rec.AddImageToDataset(pngPath, "Amy"); err != nil {
		t.Fatalf("AddImageToDataset(png, UseGray=true): %v", err)
	}
}

// TestAddImageToDatasetRejectsPNGWithoutUseGray documents the boundary of
// the above: with UseGray=false the source file is passed straight to
// go-face without any conversion, so it must already be a JPEG.
func TestAddImageToDatasetRejectsPNGWithoutUseGray(t *testing.T) {
	pngPath := filepath.Join(t.TempDir(), "amy.png")
	convertJPEGToPNG(t, filepath.Join(testFotosDir, "amy.jpg"), pngPath)

	rec := newTestRecognizer(t)
	rec.UseGray = false

	if err := rec.AddImageToDataset(pngPath, "Amy"); err == nil {
		t.Fatalf("AddImageToDataset(png, UseGray=false): expected an error, got nil")
	}
}

func convertJPEGToPNG(t *testing.T, jpegPath, pngPath string) {
	t.Helper()

	src, err := os.Open(jpegPath)
	if err != nil {
		t.Fatalf("Open(%s): %v", jpegPath, err)
	}
	defer src.Close()

	img, err := jpeg.Decode(src)
	if err != nil {
		t.Fatalf("jpeg.Decode: %v", err)
	}

	dst, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("Create(%s): %v", pngPath, err)
	}
	defer dst.Close()

	if err := png.Encode(dst, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
}

func TestRecognizeMultiples(t *testing.T) {
	rec := newTestRecognizer(t)

	faces, err := rec.RecognizeMultiples(filepath.Join(testFotosDir, "elenco3.jpg"))
	if err != nil {
		t.Fatalf("RecognizeMultiples: %v", err)
	}

	if len(faces) != 7 {
		t.Errorf("got %d faces, want 7", len(faces))
	}
}

func TestRecognizeSingleRejectsMultipleFaces(t *testing.T) {
	rec := newTestRecognizer(t)

	_, err := rec.RecognizeSingle(filepath.Join(testFotosDir, "elenco3.jpg"))
	if err == nil {
		t.Fatalf("RecognizeSingle: expected an error for a multi-face image, got nil")
	}
}

// TestAddImageToDatasetIsClassifiableWithoutSetSamples is a regression test
// for the exact limitation reported upstream in Kagami/go-face#58 and #59:
// AddImageToDataset must keep the classifier in sync incrementally (via
// goFace.AppendSample), without requiring a separate SetSamples call. It
// also exercises GrayScale's JPEG encoding path (UseGray defaults to true),
// which is a regression test in its own right -- see TestGrayScale... in
// image_test.go for the isolated version.
func TestAddImageToDatasetIsClassifiableWithoutSetSamples(t *testing.T) {
	rec := newTestRecognizer(t)
	rec.Tolerance = 0.4

	known := map[string]string{
		"Amy":        "amy.jpg",
		"Bernadette": "bernadette.jpg",
		"Howard":     "howard.jpg",
		"Penny":      "penny.jpg",
		"Raj":        "raj.jpg",
		"Sheldon":    "sheldon.jpg",
		"Leonard":    "leonard.jpg",
	}

	for id, file := range known {
		if err := rec.AddImageToDataset(filepath.Join(testFotosDir, file), id); err != nil {
			t.Fatalf("AddImageToDataset(%s): %v", id, err)
		}
	}

	// Deliberately not calling rec.SetSamples() here.
	faces, err := rec.ClassifyMultiples(filepath.Join(testFotosDir, "elenco3.jpg"))
	if err != nil {
		t.Fatalf("ClassifyMultiples: %v", err)
	}

	got := make(map[string]bool)
	for _, f := range faces {
		got[f.Id] = true

		if f.Distance <= 0 {
			t.Errorf("%s: Distance = %v, want > 0", f.Id, f.Distance)
		}
		if f.Confidence < 0 || f.Confidence > 1 {
			t.Errorf("%s: Confidence = %v, want in [0,1]", f.Id, f.Confidence)
		}
		if len(f.Shapes) == 0 {
			t.Errorf("%s: Shapes is empty, want landmark points", f.Id)
		}
	}

	for id := range known {
		if !got[id] {
			t.Errorf("%s was not classified in the group photo", id)
		}
	}
}

func TestClassifyConfidenceOfExactMatch(t *testing.T) {
	rec := newTestRecognizer(t)
	rec.Tolerance = 0.4

	if err := rec.AddImageToDataset(filepath.Join(testFotosDir, "amy.jpg"), "Amy"); err != nil {
		t.Fatalf("AddImageToDataset: %v", err)
	}

	faces, err := rec.Classify(filepath.Join(testFotosDir, "amy.jpg"))
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if len(faces) != 1 {
		t.Fatalf("got %d faces, want 1", len(faces))
	}

	f := faces[0]
	if f.Id != "Amy" {
		t.Errorf("Id = %q, want Amy", f.Id)
	}
	// The same photo classified against itself should be at (or very near)
	// zero distance, so confidence should be at (or very near) 1.
	if f.Distance > 0.01 {
		t.Errorf("Distance = %v, want close to 0 for an exact match", f.Distance)
	}
	if f.Confidence < 0.95 {
		t.Errorf("Confidence = %v, want close to 1 for an exact match", f.Confidence)
	}
}

func TestClassifyNoMatch(t *testing.T) {
	rec := newTestRecognizer(t)
	rec.Tolerance = 0.4

	if err := rec.AddImageToDataset(filepath.Join(testFotosDir, "amy.jpg"), "Amy"); err != nil {
		t.Fatalf("AddImageToDataset: %v", err)
	}

	// Raj isn't in the dataset, so classification against a lone photo of
	// him should find no match.
	_, err := rec.Classify(filepath.Join(testFotosDir, "raj.jpg"))
	if err == nil {
		t.Fatalf("Classify: expected an error for an unknown face, got nil")
	}
}

func TestSaveLoadDataset(t *testing.T) {
	rec := newTestRecognizer(t)

	if err := rec.AddImageToDataset(filepath.Join(testFotosDir, "amy.jpg"), "Amy"); err != nil {
		t.Fatalf("AddImageToDataset: %v", err)
	}

	path := filepath.Join(t.TempDir(), "dataset.json")
	if err := rec.SaveDataset(path); err != nil {
		t.Fatalf("SaveDataset: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("dataset file mode = %v, want 0600", perm)
	}

	loaded := &Recognizer{}
	if err := loaded.Init(testModelsDir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(loaded.Close)

	if err := loaded.LoadDataset(path); err != nil {
		t.Fatalf("LoadDataset: %v", err)
	}

	if len(loaded.Dataset) != 1 || loaded.Dataset[0].Id != "Amy" {
		t.Fatalf("loaded Dataset = %+v, want one entry with Id Amy", loaded.Dataset)
	}
	if loaded.Dataset[0].Descriptor != rec.Dataset[0].Descriptor {
		t.Errorf("loaded descriptor does not match the saved descriptor")
	}
}

func TestLoadDatasetMissingFile(t *testing.T) {
	rec := newTestRecognizer(t)

	err := rec.LoadDataset(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatalf("LoadDataset: expected an error for a missing file, got nil")
	}
}

// TestConcurrentAddAndClassify exercises the sync.RWMutex added around
// Dataset: run with `go test -race` to catch any regression.
func TestConcurrentAddAndClassify(t *testing.T) {
	rec := newTestRecognizer(t)
	rec.Tolerance = 0.4

	files := map[string]string{
		"Amy":        "amy.jpg",
		"Bernadette": "bernadette.jpg",
		"Howard":     "howard.jpg",
		"Penny":      "penny.jpg",
	}

	var wg sync.WaitGroup

	for id, file := range files {
		wg.Add(1)
		go func(id, file string) {
			defer wg.Done()
			if err := rec.AddImageToDataset(filepath.Join(testFotosDir, file), id); err != nil {
				t.Errorf("AddImageToDataset(%s): %v", id, err)
			}
		}(id, file)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rec.ClassifyMultiples(filepath.Join(testFotosDir, "elenco3.jpg"))
			_ = rec.SaveDataset(filepath.Join(t.TempDir(), "dataset.json"))
		}()
	}

	wg.Wait()

	if len(rec.Dataset) != len(files) {
		t.Errorf("Dataset has %d entries, want %d", len(rec.Dataset), len(files))
	}
}
