package types

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var testConfig = OllamaConfig{
	ModelFormat:   "gguf",
	ModelFamilly:  "llama",
	ModelFamilies: []string{"llama"},
	ModelType:     "7B",
	FileType:      "Q4_0",
	Architecture:  "amd64",
	OS:            "linux",
	RootFS: RootFS{
		Type: "layers",
		DiffIDs: []string{
			"sha256:e8a35b5937a5e6d5c35d1f2a15f161e07eefe5e5bb0a3cdd42998ee79b057730",
			"sha256:43070e2d4e532684de521b885f385d0841030efa2b1a20bafb76133a5e1379c1",
			"sha256:e6836092461ffbb2b06d001fce20697f62bfd759c284ee82b581ef53c55de36e",
			"sha256:ed11eda7790d05b49395598a42b155812b17e263214292f7b87d15e14003d337",
		},
	},
}

func TestOllamaConfigGeneratesDigest(t *testing.T) {
	d, err := testConfig.Digest()
	require.NoError(t, err)
	require.Equal(t, "sha256:48a8aea9cdf7af338ca5c267cb193f0b3c4b2fb9c5f2f0e0ee8b0a9dd98d1e8e", d)
}

func TestOllamaConfigWritesToDisk(t *testing.T) {
	dir := t.TempDir()

	err := testConfig.WriteToDisk(dir)
	require.NoError(t, err)

	require.FileExists(t, path.Join(dir, "sha256-48a8aea9cdf7af338ca5c267cb193f0b3c4b2fb9c5f2f0e0ee8b0a9dd98d1e8e"))
}
func TestOllamaConfigSize(t *testing.T) {

	size, err := testConfig.Size()
	require.NoError(t, err)
	require.Equal(t, 484, size)
}

var kParams = map[string][]string{
	"mirostat":       []string{"2"},
	"mirostat_eta":   []string{"0.1"},
	"mirostat_tau":   []string{"5.0"},
	"num_ctx":        []string{"4096"},
	"repeat_last_n":  []string{"64"},
	"repeat_penalty": []string{"1.1"},
	"temperature":    []string{"0.7"},
	"seed":           []string{"42"},
	"stop":           []string{"[a]", "[b]"},
	"tfs_z":          []string{"1.1"},
	"num_predict":    []string{"23"},
	"top_k":          []string{"42"},
	"top_p":          []string{"0.7"},
}

func TestConvertsParametersCorrectly(t *testing.T) {
	// convert the test collection to a json string wrapped in a zipped
	// reader as it will be received from the layer
	w := bytes.Buffer{}
	gzw := gzip.NewWriter(&w)
	err := json.NewEncoder(gzw).Encode(kParams)
	require.NoError(t, err)
	gzw.Close()

	// create an io.Reader from the zipped and encoded data
	reader := io.NopCloser(bytes.NewReader(w.Bytes()))

	// convert params
	out := ConvertKapsuleParamsToOllamaParams(reader)

	// convert the output back into a collection for testing
	oParams := map[string]interface{}{}
	err = json.NewDecoder(out).Decode(&oParams)
	require.NoError(t, err)

	// eval
	require.Equal(t, float64(2), oParams["mirostat"])
	require.Equal(t, 0.1, oParams["mirostat_eta"])
	require.Equal(t, []interface{}{"[a]", "[b]"}, oParams["stop"])
}
