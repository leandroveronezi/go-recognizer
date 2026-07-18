## go-recognizer

Face detection and recognition for Go, built on top of [dlib](http://dlib.net)
via [go-face](https://github.com/leandroveronezi/go-face). It wraps the
lower-level go-face API into a small, batteries-included `Recognizer` type:
load a photo, find faces, classify them against a labeled dataset, and draw
the results back onto the image — in a handful of method calls.

[![CI](https://github.com/leandroveronezi/go-recognizer/actions/workflows/ci.yml/badge.svg)](https://github.com/leandroveronezi/go-recognizer/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/leandroveronezi/go-recognizer.svg)](https://pkg.go.dev/github.com/leandroveronezi/go-recognizer)
[![Latest tag](https://img.shields.io/github/v/tag/leandroveronezi/go-recognizer.svg)](https://github.com/leandroveronezi/go-recognizer/tags)
![MIT Licensed](https://img.shields.io/github/license/leandroveronezi/go-recognizer.svg)
[![](https://img.shields.io/badge/Require-go--face-blue.svg)](https://github.com/leandroveronezi/go-face)

## Features

- **Detection** — find one or many faces in an image, sorted left to right.
- **Recognition** — classify detected faces against a dataset of known people.
- **Incremental dataset updates** — `AddImageToDataset` keeps the classifier in
  sync as each face is added; no need to rebuild the whole sample set.
- **Match distance/confidence** — `Classify`/`ClassifyMultiples` return the
  matched face's `Distance` and a normalized `Confidence` score, not just an ID.
- **Landmarks** — detected faces carry their `Shapes` (facial landmark
  points). Defaults to 5 points (eye corners, nose base); set
  `rec.Model.Landmark` before `Init` to opt into the 68-point model for
  full facial contour (jawline, eyebrows, nose bridge, eyes, lips).
- **Swappable model files** — `rec.Model.Landmark`/`Descriptor`/`CNN` let
  you point Init at differently-named or fine-tuned model files instead
  of go-face's defaults.
- **Configurable matching** — tune the distance `Tolerance` used to accept a match.
- **CNN or HOG detector** — trade speed for accuracy with `UseCNN`.
- **Grayscale preprocessing** — optional, via `UseGray`.
- **Beyond JPEG input** — go-face's own file loader only understands JPEG,
  but with the default `UseGray = true`, go-recognizer decodes the source
  image with Go's standard `image` package first (JPEG and PNG are
  supported out of the box) and re-encodes it before handing it to
  go-face, so PNG sources work without extra steps. This doesn't apply
  when `UseGray = false`: the original file is passed straight through,
  so it must already be a JPEG.
- **Dataset persistence** — save/load known faces to/from a JSON file.
- **Drawing helpers** — annotate the source image with boxes, labels, and
  landmark points for the faces found.

## Requirements

go-recognizer depends on go-face, which in turn requires dlib (>= 19.10) and the
libjpeg development headers to compile.

go-face uses cgo, so `CGO_ENABLED=1` is required at build time (this is the
default on most setups, but some environments/CI images turn it off). If you
see errors like `undefined: face.NewRecognizer` or `undefined: face.Descriptor`
instead of a compiler error, that's almost always CGO being disabled — run
`go env -w CGO_ENABLED=1` or set the env var for the build.

### Ubuntu 18.10+, Debian sid

Latest versions of Ubuntu and Debian provide a suitable dlib package, so just run:

```bash
# Ubuntu
sudo apt-get install libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg-turbo8-dev
# Debian
sudo apt-get install libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg62-turbo-dev
```

### macOS

Make sure you have [Homebrew](https://brew.sh) installed.

```bash
brew install dlib
```

### Windows

Make sure you have [MSYS2](https://www.msys2.org) installed.

1. Run `MSYS2 MSYS` shell from Start menu
2. Run `pacman -Syu` and if it asks you to close the shell do that
3. Run `pacman -Syu` again
4. Run `pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-dlib`
5.
   1. If you already have Go and Git installed and available in PATH uncomment
      `set MSYS2_PATH_TYPE=inherit` line in `msys2_shell.cmd` located in MSYS2
      installation folder
   2. Otherwise run `pacman -S mingw-w64-x86_64-go git`
6. Run `MSYS2 MinGW 64-bit` shell from Start menu to compile and use go-face

### Other systems

Try installing dlib/libjpeg with your distribution's package manager, or
[compile dlib from source](http://dlib.net/compile.html). go-face won't work
with old dlib packages such as libdlib18. If your system isn't covered here,
open an issue with the distribution/version and we'll try to help.

### Docker

[`examples/Dockerfile`](examples/Dockerfile) builds dlib from source (Alpine
has no dlib package) and compiles the detection example against it, ending
with a ~50MB runtime image. Useful as a reference for containerized builds,
and for the compiler/CMake compatibility patches it applies -- dlib's
released source doesn't build out of the box with GCC 15+ or CMake 4.x.

## Installation

```bash
go get github.com/leandroveronezi/go-recognizer
```

```go
import "github.com/leandroveronezi/go-recognizer"
```

## Models

`shape_predictor_5_face_landmarks.dat`, `mmod_human_face_detector.dat` and
`dlib_face_recognition_resnet_model_v1.dat` are required at runtime. Download
them from the [dlib-models](https://github.com/davisking/dlib-models) repo:

```bash
mkdir models && cd models
wget https://github.com/davisking/dlib-models/raw/master/shape_predictor_5_face_landmarks.dat.bz2
bunzip2 shape_predictor_5_face_landmarks.dat.bz2
wget https://github.com/davisking/dlib-models/raw/master/dlib_face_recognition_resnet_model_v1.dat.bz2
bunzip2 dlib_face_recognition_resnet_model_v1.dat.bz2
wget https://github.com/davisking/dlib-models/raw/master/mmod_human_face_detector.dat.bz2
bunzip2 mmod_human_face_detector.dat.bz2
```

Optional: `shape_predictor_68_face_landmarks.dat` for full facial contour
landmarks (see `rec.Model.Landmark` below). It's a much larger download
(~95MB uncompressed, vs ~9MB for the 5-point model), so it's opt-in rather
than required.

```bash
wget https://github.com/davisking/dlib-models/raw/master/shape_predictor_68_face_landmarks.dat.bz2
bunzip2 shape_predictor_68_face_landmarks.dat.bz2
```

## Examples

Runnable versions of the examples below live in [`examples/`](examples/), one
per subfolder. Run them from inside `examples/` so the relative `fotos`/`models`
paths resolve, e.g. `cd examples && go run ./detection`.

#### Face detection

```go
package main

import (
	"fmt"
	"path/filepath"

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

	rec.SaveImage("faces2.jpg", img)

}
```

![Face detection result](https://leandroveronezi.github.io/go-recognizer/examples/faces2.jpg)

#### Face recognition

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/leandroveronezi/go-recognizer"
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

	// No rec.SetSamples() call needed here: AddImageToDataset already
	// keeps the classifier in sync incrementally as each face is added.

	faces, err := rec.ClassifyMultiples(filepath.Join(fotosDir, "elenco3.jpg"))

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
```

![Face recognition result](https://leandroveronezi.github.io/go-recognizer/examples/faces.jpg)

#### Face landmarks

```go
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
```

![Face landmarks result](https://leandroveronezi.github.io/go-recognizer/examples/landmarks.jpg)

This uses the default 5-point model. For the full facial contour, download
`shape_predictor_68_face_landmarks.dat` (see [Models](#models)) and set
`rec.Model.Landmark` before `Init`:

```go
rec := recognizer.Recognizer{}
rec.Model.Landmark = "shape_predictor_68_face_landmarks.dat"
err := rec.Init(dataDir)
```

`rec.Model.Descriptor` and `rec.Model.CNN` work the same way, for the
face-descriptor (ResNet) and CNN detector model files respectively —
useful if you're using differently-named or fine-tuned dlib models. All
three must be set before calling `Init`; they're read once, at load time.

## Contributing

Issues and pull requests are welcome. If you're reporting a build problem,
please include your OS/distribution, Go version, and the full compiler
output.

## License

[MIT](LICENSE)
