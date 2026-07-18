package recognizer

import (
	"encoding/json"
	"errors"
	"os"
)

/*
SaveDataset saves Dataset data to a json file
*/
func (_this *Recognizer) SaveDataset(Path string) error {

	data, err := func() ([]byte, error) {
		_this.mu.RLock()
		defer _this.mu.RUnlock()
		return jsonMarshal(_this.Dataset)
	}()

	if err != nil {
		return err
	}

	if err := os.WriteFile(Path, data, 0600); err != nil {
		return err
	}

	return os.Chmod(Path, 0600)

}

/*
LoadDataset loads the data from the json file into the Dataset.

Call SetSamples afterward: as with AddImageToDataset, Classify and
ClassifyMultiples won't see the loaded entries until SetSamples runs
again.
*/
func (_this *Recognizer) LoadDataset(Path string) error {

	if !fileExists(Path) {
		return errors.New("file not found")
	}

	file, err := os.Open(Path)

	if err != nil {
		return err
	}
	defer file.Close()

	Dataset := make([]Data, 0)
	err = json.NewDecoder(file).Decode(&Dataset)
	if err != nil {
		return err
	}

	_this.mu.Lock()
	_this.Dataset = append(_this.Dataset, Dataset...)
	_this.mu.Unlock()

	return nil

}
