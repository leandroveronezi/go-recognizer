package recognizer

import (
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// defaultModelFiles are the model files Init (with the default,
// unconfigured Model) requires.
var defaultModelFiles = []string{
	"shape_predictor_5_face_landmarks.dat",
	"dlib_face_recognition_resnet_model_v1.dat",
	"mmod_human_face_detector.dat",
}

var modelsHTTPClient = &http.Client{Timeout: 5 * time.Minute}

/*
DownloadModels downloads the model files Init requires by default
(shape_predictor_5_face_landmarks.dat, dlib_face_recognition_resnet_model_v1.dat,
mmod_human_face_detector.dat) from the dlib-models repository into dir,
creating dir if it doesn't exist yet.

Any file that already exists in dir is left untouched and not
re-downloaded, so it's safe to call this on every startup before Init.
*/
func (_this *Recognizer) DownloadModels(dir string) error {

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	for _, name := range defaultModelFiles {

		path := filepath.Join(dir, name)

		if fileExists(path) {
			continue
		}

		if err := downloadModel(path, name); err != nil {
			return err
		}

	}

	return nil

}

// downloadModel fetches the bunzip2-compressed model file "name" from the
// dlib-models repository and writes the decompressed contents to path.
// path is left absent (not a partial file) if anything fails partway.
func downloadModel(path, name string) error {

	url := "https://github.com/davisking/dlib-models/raw/master/" + name + ".bz2"

	resp, err := modelsHTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: unexpected status %s", name, resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, bzip2.NewReader(resp.Body)); err != nil {
		out.Close()
		os.Remove(path)
		return err
	}

	return out.Close()

}
