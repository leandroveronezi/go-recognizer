package recognizer

import (
	goFace "github.com/leandroveronezi/go-face"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

/*
LoadImage Load an image from file
*/
func (_this *Recognizer) LoadImage(Path string) (image.Image, error) {

	existingImageFile, err := os.Open(Path)
	if err != nil {
		return nil, err
	}
	defer existingImageFile.Close()

	imageData, _, err := image.Decode(existingImageFile)
	if err != nil {
		return nil, err
	}

	return imageData, nil

}

/*
SaveImage Save an image to jpeg file
*/
func (_this *Recognizer) SaveImage(Path string, Img image.Image) error {

	outputFile, err := os.Create(Path)
	if err != nil {
		return err
	}

	err = jpeg.Encode(outputFile, Img, nil)

	if err != nil {
		outputFile.Close()
		return err
	}

	return outputFile.Close()

}

/*
GrayScale Convert an image to grayscale
*/
func (_this *Recognizer) GrayScale(imgSrc image.Image) image.Image {

	// NRGBA, not image.Gray: jpeg.Encode special-cases *image.Gray into a
	// single-component JPEG, which go-face's jpeg_mem_loader rejects
	// outright (it requires exactly 3 components). Converting each pixel
	// to gray but storing it in an NRGBA image keeps the visual result
	// grayscale while still encoding as a 3-component JPEG.
	bounds := imgSrc.Bounds()
	gray := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, color.GrayModel.Convert(imgSrc.At(x, y)))
		}
	}

	return gray

}

/*
loadPixels loads Path, converts it to grayscale, and returns it as a
tightly packed RGB pixel buffer (3 bytes per pixel, row-major) ready for
goFace's Raw recognition methods -- no JPEG encoding/decoding involved.
*/
func (_this *Recognizer) loadPixels(Path string) (pixels []byte, width, height int, err error) {

	img, err := _this.LoadImage(Path)

	if err != nil {
		return nil, 0, 0, err
	}

	img = _this.GrayScale(img)

	bounds := img.Bounds()
	width, height = bounds.Dx(), bounds.Dy()
	pixels = make([]byte, width*height*3)

	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixels[i] = byte(r >> 8)
			pixels[i+1] = byte(g >> 8)
			pixels[i+2] = byte(b >> 8)
			i += 3
		}
	}

	return pixels, width, height, nil

}

/*
DrawFaces draws the faces identified in the original image
*/
func (_this *Recognizer) DrawFaces(Path string, F []Face) (image.Image, error) {

	img, err := _this.LoadImage(Path)

	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(font, &truetype.Options{Size: 24})

	dc := gg.NewContextForImage(img)
	dc.SetFontFace(face)

	for _, f := range F {

		dc.SetRGB255(0, 0, 255)

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

	return img, nil

}

/*
DrawFaces2 draws the faces in the original image
*/
func (_this *Recognizer) DrawFaces2(Path string, F []goFace.Face) (image.Image, error) {

	aux := make([]Face, 0)

	for _, f := range F {

		auxFace := Face{}
		auxFace.Rectangle = f.Rectangle
		auxFace.Descriptor = f.Descriptor
		auxFace.Shapes = f.Shapes

		aux = append(aux, auxFace)

	}

	return _this.DrawFaces(Path, aux)

}

/*
DrawLandmarks draws the facial landmark points found in the original image.
*/
func (_this *Recognizer) DrawLandmarks(Path string, F []goFace.Face) (image.Image, error) {

	img, err := _this.LoadImage(Path)

	if err != nil {
		return nil, err
	}

	dc := gg.NewContextForImage(img)
	dc.SetRGB255(0, 255, 0)

	for _, f := range F {
		for _, p := range f.Shapes {
			dc.DrawPoint(float64(p.X), float64(p.Y), 3)
		}
	}

	dc.Fill()

	img = dc.Image()

	return img, nil

}
