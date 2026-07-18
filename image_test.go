package recognizer

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"path/filepath"
	"testing"

	goFace "github.com/leandroveronezi/go-face"
)

// TestGrayScaleEncodesAsThreeComponentJPEG is a regression test for a real
// bug: GrayScale used to build an *image.Gray, and image/jpeg.Encode
// special-cases that concrete type into a single-component JPEG, which
// go-face's jpeg_mem_loader rejects outright ("jpeg_mem_loader: unsupported
// pixel size"). That broke AddImageToDataset/RecognizeFile whenever UseGray
// was true -- i.e. always, by default. This test needs no dlib/model files,
// so it catches the regression even without running the heavier
// model-backed tests.
func TestGrayScaleEncodesAsThreeComponentJPEG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src.Set(x, y, color.RGBA{R: 200, G: 50, B: 10, A: 255})
		}
	}

	rec := &Recognizer{}
	gray := rec.GrayScale(src)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, gray, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}

	decoded, err := jpeg.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("jpeg.Decode: %v", err)
	}

	if _, ok := decoded.(*image.Gray); ok {
		t.Fatalf("re-decoded image is *image.Gray (single-component JPEG); go-face's jpeg_mem_loader requires 3 components")
	}

	b := decoded.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := decoded.At(x, y).RGBA()
			if r != g || g != bl {
				t.Fatalf("pixel (%d,%d) is not grayscale: r=%d g=%d b=%d", x, y, r, g, bl)
			}
		}
	}
}

func TestDrawFacesAndDrawFaces2(t *testing.T) {
	rec := &Recognizer{}
	facesPath := filepath.Join(testFotosDir, "amy.jpg")

	faces := []Face{
		{Data: Data{Id: "Test"}, Rectangle: image.Rect(10, 10, 50, 50)},
	}
	img, err := rec.DrawFaces(facesPath, faces)
	if err != nil {
		t.Fatalf("DrawFaces: %v", err)
	}
	if img.Bounds().Empty() {
		t.Fatalf("DrawFaces returned an empty image")
	}

	goFaces := []goFace.Face{
		goFace.New(image.Rect(10, 10, 50, 50), goFace.Descriptor{}),
	}
	img2, err := rec.DrawFaces2(facesPath, goFaces)
	if err != nil {
		t.Fatalf("DrawFaces2: %v", err)
	}
	if img2.Bounds() != img.Bounds() {
		t.Errorf("DrawFaces2 bounds = %v, want %v", img2.Bounds(), img.Bounds())
	}
}

func TestDrawLandmarks(t *testing.T) {
	rec := &Recognizer{}
	facesPath := filepath.Join(testFotosDir, "amy.jpg")

	goFaces := []goFace.Face{
		goFace.NewWithShape(
			image.Rect(10, 10, 50, 50),
			[]image.Point{{X: 20, Y: 20}, {X: 30, Y: 30}},
			goFace.Descriptor{},
		),
	}

	img, err := rec.DrawLandmarks(facesPath, goFaces)
	if err != nil {
		t.Fatalf("DrawLandmarks: %v", err)
	}
	if img.Bounds().Empty() {
		t.Fatalf("DrawLandmarks returned an empty image")
	}
}

func TestSaveImageAndLoadImageRoundTrip(t *testing.T) {
	rec := &Recognizer{}

	src := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.Set(x, y, color.RGBA{R: uint8(x * 10), G: uint8(y * 10), B: 100, A: 255})
		}
	}

	path := filepath.Join(t.TempDir(), "roundtrip.jpg")
	if err := rec.SaveImage(path, src); err != nil {
		t.Fatalf("SaveImage: %v", err)
	}

	loaded, err := rec.LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}

	if loaded.Bounds() != src.Bounds() {
		t.Errorf("loaded bounds = %v, want %v", loaded.Bounds(), src.Bounds())
	}
}
