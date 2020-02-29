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

To use go-face in your Go code:

```go
import "github.com/leandroveronezi/go-recognizer"
```

To install go-face in your $GOPATH:

```bash
go get github.com/leandroveronezi/go-recognizer
```


## Examples

###### Face detection 

```go
package main

import (
	"fmt"
	"github.com/leandroveronezi/recognizer"
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

	if err == nil {
		rec.SaveImage("faces2.jpeg", img)
	}

}
```

![](https://leandroveronezi.github.io/go-recognizer/examples/faces2.jpg)









###### Face recognition 

```go
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

	if err == nil {
		rec.SaveImage("faces.jpg", img)
	}

}

```

![](https://leandroveronezi.github.io/go-recognizer/examples/faces.jpg)