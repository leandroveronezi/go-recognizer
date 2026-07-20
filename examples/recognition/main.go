package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/leandroveronezi/go-recognizer"
)

const fotosDir = "fotos"
const dataDir = "models"

func addFile(rec *recognizer.Recognizer, Path, Id string) {

	err := rec.AddImageToDataset(Path, Id)

	switch {
	case errors.Is(err, recognizer.ErrNoFace), errors.Is(err, recognizer.ErrNotSingleFace):
		fmt.Printf("%s: not exactly one face, skipping\n", Path)
	case err != nil:
		fmt.Println(err)
	}

}

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

	addFile(&rec, filepath.Join(fotosDir, "amy.jpg"), "Amy")
	addFile(&rec, filepath.Join(fotosDir, "bernadette.jpg"), "Bernadette")
	addFile(&rec, filepath.Join(fotosDir, "howard.jpg"), "Howard")
	addFile(&rec, filepath.Join(fotosDir, "penny.jpg"), "Penny")
	addFile(&rec, filepath.Join(fotosDir, "raj.jpg"), "Raj")
	addFile(&rec, filepath.Join(fotosDir, "sheldon.jpg"), "Sheldon")
	addFile(&rec, filepath.Join(fotosDir, "leonard.jpg"), "Leonard")

	// No rec.SetSamples() call needed here: AddImageToDataset already
	// keeps the classifier in sync incrementally as each face is added.

	faces, err := rec.IdentifyMultiples(filepath.Join(fotosDir, "elenco3.jpg"))

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, f := range faces {
		fmt.Printf("%s: distance=%.4f confidence=%.2f%%\n", f.Id, f.Distance, f.Confidence*100)
	}

	img, err := rec.DrawFaces(filepath.Join(fotosDir, "elenco3.jpg"), faces)

	if err != nil {
		fmt.Println(err)
		return
	}

	rec.SaveImage("faces.jpg", img)

}
