## Face detection and recognition using golang

[![Go Report Card](https://goreportcard.com/badge/github.com/leandroveronezi/go-recognizer)](https://goreportcard.com/report/github.com/leandroveronezi/go-recognizer)
[![GoDoc](https://godoc.org/github.com/leandroveronezi/go-recognizer?status.png)](https://godoc.org/github.com/leandroveronezi/go-recognizer)
![MIT Licensed](https://img.shields.io/github/license/leandroveronezi/go-recognizer.svg)
![](https://img.shields.io/github/repo-size/leandroveronezi/go-recognizer.svg)
[![](https://img.shields.io/badge/Require-go--face-blue.svg)](https://github.com/Kagami/go-face)

## Requirements
go-recognizer require go-face to compile. go-face need to have dlib (>= 19.10) and libjpeg development packages installed.
For install details see [go-face](https://github.com/Kagami/go-face) documentation.

## Usage

To use go-recognizer in your Go code:

```go
import "github.com/leandroveronezi/go-recognizer"
```

To install go-recognizer in your $GOPATH:

```bash
go get github.com/leandroveronezi/go-recognizer
```

## Models

Currently `shape_predictor_5_face_landmarks.dat`, `mmod_human_face_detector.dat` and
`dlib_face_recognition_resnet_model_v1.dat` are required. You may download them
from [dlib-models](https://github.com/davisking/dlib-models) repo:

```bash
mkdir models && cd models
wget https://github.com/davisking/dlib-models/raw/master/shape_predictor_5_face_landmarks.dat.bz2
bunzip2 shape_predictor_5_face_landmarks.dat.bz2
wget https://github.com/davisking/dlib-models/raw/master/dlib_face_recognition_resnet_model_v1.dat.bz2
bunzip2 dlib_face_recognition_resnet_model_v1.dat.bz2
wget https://github.com/davisking/dlib-models/raw/master/mmod_human_face_detector.dat.bz2
bunzip2 mmod_human_face_detector.dat.bz2
```

## Examples

###### Face detection 

```go
package main

import (
	"fmt"
	"github.com/leandroveronezi/go-recognizer"
	"path/filepath"
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

	faces, err := rec.RecognizeMultiples(filepath.Join(fotosDir, "elenco3.jpg"))

	if err != nil {
		fmt.Println(err)
		return
	}

	img, err := rec.DrawFaces2(filepath.Join(fotosDir, "elenco3.jpg"), faces)

    	if err != nil {
		fmt.Println(err)
		return
	}
	
	rec.SaveImage("faces2.jpeg", img)

}
```

![](https://leandroveronezi.github.io/go-recognizer/examples/faces2.jpg)









###### Face recognition 

```go
package main

import (
	"fmt"
	"github.com/leandroveronezi/go-recognizer"
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

	rec.SetSamples()

	faces, err := rec.ClassifyMultiples(filepath.Join(fotosDir, "elenco3.jpg"))

	if err != nil {
		fmt.Println(err)
		return
	}

	img, err := rec.DrawFaces(filepath.Join(fotosDir, "elenco3.jpg"), faces)

    	if err != nil {
		fmt.Println(err)
		return
	}
	
	rec.SaveImage("faces.jpg", img)

}

```

![](https://leandroveronezi.github.io/go-recognizer/examples/faces.jpg)
