package types

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const OLLAMA_MEDIA_TYPE_MODEL = "application/vnd.ollama.image.model"
const OLLAMA_MEDIA_TYPE_LICENCE = "application/vnd.ollama.image.licence"
const OLLAMA_MEDIA_TYPE_TEMPLATE = "application/vnd.ollama.image.template"
const OLLAMA_MEDIA_TYPE_PARAMETERS = "application/vnd.ollama.image.params"

// OllamaConfig is the docker manifest config for the image
type OllamaConfig struct {
	ModelFormat   string   `json:"model_format"`
	ModelFamilly  string   `json:"model_familly"`
	ModelFamilies []string `json:"model_famillies"`
	ModelType     string   `json:"model_type"`
	FileType      string   `json:"file_type"`
	Architecture  string   `json:"architecture"`
	OS            string   `json:"os"`
	RootFS        RootFS   `json:"rootfs"`
}

// RootFS defines the image digests that are referenced by the config
type RootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

func (o OllamaConfig) Size() (int, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config to JSON: %s", err)
	}
	return len(data), nil
}

// WriteToDisk writes the config to a file in the blobs
// folder. The file will have the name of the digest
func (o OllamaConfig) WriteToDisk(blobs string) error {
	data, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %s", err)
	}

	digest, err := o.Digest()
	if err != nil {
		return err
	}

	// replace the digest name
	digest = strings.Replace(digest, ":", "-", 1)
	fp := path.Join(blobs, digest)

	os.Remove(fp)

	err = os.WriteFile(fp, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to disk: %s", err)
	}

	return nil
}

func (o OllamaConfig) Digest() (string, error) {
	d, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("unable to marshal config to json: %s", err)
	}

	hash := sha256.Sum256(d)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(hash[0:])), nil
}

var ollamaParametersDict = map[string]string{
	"mirostat":       "int",
	"mirostat_eta":   "float",
	"mirostat_tau":   "float",
	"num_ctx":        "int",
	"repeat_last_n":  "int",
	"repeat_penalty": "float",
	"temperature":    "float",
	"seed":           "int",
	"stop":           "[]string",
	"tfs_z":          "float",
	"num_predict":    "int",
	"top_k":          "int",
	"top_p":          "float",
}

// ConvertKapsuleParamsToOllamaParams converts a compressed layer containing
// a Kapsule parameter collection into the json format that is expected by ollama
// returns a writer that can be added to a new image later
func ConvertKapsuleParamsToOllamaParams(r io.ReadCloser) io.ReadCloser {
	// the layer is compressed, get a gzip reader to decompress as we write it
	gzrc, err := gzip.NewReader(r)
	if err != nil {
		return nil
	}

	// convert the reader containing json version of the params to map
	params := map[string][]string{}
	err = json.NewDecoder(gzrc).Decode(&params)
	if err != nil {
		return nil
	}

	ret := map[string]interface{}{}

	for k, v := range params {
		t := ollamaParametersDict[k]

		switch t {
		case "int":
			nv, err := convertToInt(v)
			if err == nil {
				ret[k] = nv
			}
		case "[]string":
			ret[k] = v
		case "float":
			nv, err := convertToFloat(v)
			if err == nil {
				ret[k] = nv
			}
		}
	}

	// serialize to json
	d, err := json.Marshal(&ret)
	if err != nil {
		return nil
	}

	return io.NopCloser(bytes.NewBuffer(d))
}

func convertToInt(value []string) (int, error) {
	if len(value) == 0 {
		return 0, fmt.Errorf("invalid value")
	}

	valueInt, err := strconv.Atoi(value[0])
	if err != nil {
		// handle error
		return 0, err
	}
	return valueInt, err
}

// convertStringToFloat converts a string to a float64
func convertToFloat(value []string) (float64, error) {
	if len(value) == 0 {
		return 0, fmt.Errorf("invalid value")
	}

	floatValue, err := strconv.ParseFloat(value[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to float: %s", err)
	}
	return floatValue, nil
}
