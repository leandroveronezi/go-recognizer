package recognizer

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

func (_this *Recognizer) LoadImage(Path string) (error, image.Image) {

	existingImageFile, err := os.Open(Path)
	if err != nil {
		return err, nil
	}
	defer existingImageFile.Close()

	imageData, _, err := image.Decode(existingImageFile)
	if err != nil {
		return err, nil
	}

	return nil, imageData

}

func (_this *Recognizer) SaveImage(Path string, Img image.Image) error {

	outputFile, err := os.Create(Path)
	if err != nil {
		return err
	}

	err = jpeg.Encode(outputFile, Img, nil)

	if err != nil {
		return err
	}

	return outputFile.Close()

}

func (_this *Recognizer) GrayScale(imgSrc image.Image) image.Image {

	return imaging.Grayscale(imgSrc)

}

func (_this *Recognizer) createTempGrayFile(Path, Id string) (error, string) {

	name := _this.TempFileName(Id, ".jpeg")

	err, img := _this.LoadImage(Path)

	if err != nil {
		return err, ""
	}

	img = _this.GrayScale(img)
	err = _this.SaveImage(name, img)

	if err != nil {
		return err, ""
	}

	return nil, name

}

// TempFileName generates a temporary filename for use in testing or whatever
func (_this *Recognizer) TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func (_this *Recognizer) DrawFaces(Path string, F []Face) (error, image.Image) {

	err, img := _this.LoadImage(Path)

	if err != nil {
		return err, nil
	}

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}

	face := truetype.NewFace(font, &truetype.Options{Size: 24})

	dc := gg.NewContextForImage(img)
	dc.SetFontFace(face)

	for _, f := range F {

		dc.SetRGB255(50, 168, 82)

		x := float64(f.Rectangle.Min.X)
		y := float64(f.Rectangle.Min.Y)
		w := float64(f.Rectangle.Dx())
		h := float64(f.Rectangle.Dy())

		dc.DrawString(f.Id, x, y+h+20)

		dc.DrawRectangle(x, y, w, h)
		dc.SetLineWidth(4.0)
		dc.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: 0, G: 0, B: 255, A: 255}))
		dc.Stroke()

	}

	img = dc.Image()

	return nil, img

}
