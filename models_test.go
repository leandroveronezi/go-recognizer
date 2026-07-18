package recognizer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDownloadModels downloads the real model files and confirms they
// actually work with Init/RecognizeSingle -- not just that files with the
// right names landed on disk.
func TestDownloadModels(t *testing.T) {

	dir := t.TempDir()

	rec := &Recognizer{}
	if err := rec.DownloadModels(dir); err != nil {
		t.Fatalf("DownloadModels: %v", err)
	}

	for _, name := range defaultModelFiles {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("Stat(%s): %v", name, err)
		}
		if info.Size() == 0 {
			t.Errorf("%s downloaded as an empty file", name)
		}
	}

	loaded := &Recognizer{}
	if err := loaded.Init(dir); err != nil {
		t.Fatalf("Init with downloaded models: %v", err)
	}
	defer loaded.Close()

	if _, err := loaded.RecognizeSingle(filepath.Join(testFotosDir, "amy.jpg")); err != nil {
		t.Fatalf("RecognizeSingle with downloaded models: %v", err)
	}

}

// TestDownloadModelsSkipsExisting proves DownloadModels genuinely skips
// files that already exist -- it doesn't just check-then-overwrite. Uses
// sentinel files instead of real models, so it needs no network access.
func TestDownloadModelsSkipsExisting(t *testing.T) {

	dir := t.TempDir()

	sentinels := make(map[string][]byte, len(defaultModelFiles))
	for i, name := range defaultModelFiles {
		data := []byte{byte(i), 'n', 'o', 't', ' ', 'a', ' ', 'r', 'e', 'a', 'l', ' ', 'm', 'o', 'd', 'e', 'l'}
		sentinels[name] = data
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatalf("WriteFile(%s): %v", name, err)
		}
	}

	rec := &Recognizer{}
	if err := rec.DownloadModels(dir); err != nil {
		t.Fatalf("DownloadModels: %v", err)
	}

	for name, want := range sentinels {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", name, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s was overwritten, want it left untouched since it already existed", name)
		}
	}

}
