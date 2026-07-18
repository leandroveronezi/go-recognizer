package main

import (
	"fmt"
	"path/filepath"

	face "github.com/leandroveronezi/go-face"
	"github.com/leandroveronezi/go-recognizer"
)

const fotosDir = "fotos"
const dataDir = "models"

func main() {

	rec := recognizer.Recognizer{}
	err := rec.Init(dataDir)

	if err != nil {
		fmt.Println(err)
		return
	}

	rec.Tolerance = 0.4
	rec.UseGray = true
	rec.UseCNN = false
	defer rec.Close()

	f, err := rec.RecognizeSingle(filepath.Join(fotosDir, "amy.jpg"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("found %d landmark points\n", len(f.Shapes))

	img, err := rec.DrawLandmarks(filepath.Join(fotosDir, "amy.jpg"), []face.Face{f})

	if err != nil {
		fmt.Println(err)
		return
	}

	rec.SaveImage("landmarks.jpg", img)

}
