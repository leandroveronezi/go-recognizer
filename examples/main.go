package main

import (
	"fmt"
	"github.com/leandroveronezi/recognizer"
	"path/filepath"
)

const fotosDir = "fotos"
const dataDir = "models"

func addFile(rec *recognizer.Recognizer, Path, Id string) {

	err := rec.AddImageToDataset(Path, Id)

	if err != nil {
		fmt.Println(err)
		return
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
	defer rec.Close()

	addFile(&rec, filepath.Join(fotosDir, "amy.jpg"), "Amy")
	addFile(&rec, filepath.Join(fotosDir, "bernadette.jpg"), "Bernadette")
	addFile(&rec, filepath.Join(fotosDir, "howard.jpg"), "Howard")
	addFile(&rec, filepath.Join(fotosDir, "penny.jpg"), "Penny")
	addFile(&rec, filepath.Join(fotosDir, "raj.jpg"), "Raj")
	addFile(&rec, filepath.Join(fotosDir, "sheldon.jpg"), "Sheldon")
	addFile(&rec, filepath.Join(fotosDir, "leonard.jpg"), "Leonard")

	rec.LoadSamples()

	err, _, faces := rec.ClassifyMultiples(filepath.Join(fotosDir, "elenco3.jpg"))

	if err != nil {
		fmt.Println(err)
		return
	}

	err, img := rec.DrawFaces(filepath.Join(fotosDir, "elenco3.jpg"), faces)

	if err == nil {
		//img = rec.GrayScale(img)
		rec.SaveImage("teste.jpeg", img)
	}

}
