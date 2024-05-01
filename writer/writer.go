package writer

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/containers/image/v5/manifest"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/nicholasjackson/kapsule/types"
	"github.com/opencontainers/go-digest"
)

type Writer interface {
	// Write to path writes the image to a local OCI image registry defined by output
	// if no existing regitry exists at the output path, WriteToPath scaffolds a new
	// regitstry before writing the image
	WriteToPath(image v1.Image, output string) error
	// PushToRegistry pushes the given image to a remote OCI image registry
	PushToRegistry(imageRef string, image v1.Image, username, password string) error
}

// WriterImpl is a concrete implementation of the Writer interface
type WriterImpl struct{}

func WriteToPath(image v1.Image, output string) error {
	var err error
	var p layout.Path

	p, err = layout.FromPath(output)
	if err != nil {
		// no index exists at the path, create a new index
		p, err = layout.Write(output, empty.Index)
		if err != nil {
			return err
		}
	}

	err = p.AppendImage(image)
	if err != nil {
		return err
	}

	return nil
}

func PushToRegistry(imageRef string, image v1.Image, username, password string) error {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		panic(err)
	}

	b := authn.Basic{
		Username: username,
		Password: password,
	}

	cfg, err := b.Authorization()
	if err != nil {
		return err
	}

	auth := authn.FromConfig(*cfg)

	return remote.Write(ref, image, remote.WithAuth(auth))
}

func WriteToOllama(image v1.Image, imageRef, output string) error {
	cn := types.CanonicalRef(imageRef)
	ref, err := name.ParseReference(cn)
	if err != nil {
		panic(err)
	}

	manifestFolder := path.Join(output, "manifests", ref.Context().RegistryStr(), ref.Context().RepositoryStr())
	blobsFolder := path.Join(output, "blobs")

	// create the folders
	err = os.MkdirAll(manifestFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create manifests folder: %s", err)
	}

	os.MkdirAll(blobsFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create blobs folder: %s", err)
	}

	// add the layer digests
	layers, err := image.Layers()
	if err != nil {
		return fmt.Errorf("unable to read layers: %s", err)
	}

	// add the layers
	schemaLayers := []manifest.Schema2Descriptor{}
	for _, l := range layers {
		mt, err := l.MediaType()
		if err != nil {
			return fmt.Errorf("unable to get media type from layer: %s", err)
		}

		switch mt {
		case types.KAPSULE_MEDIA_TYPE_PARAMETERS:
			// handle params differently as we need to convert
			in, err := l.Compressed()
			if err != nil {
				return fmt.Errorf("unable to read layer: %w", err)
			}

			out := types.ConvertKapsuleParamsToOllamaParams(in)
			if out == nil {
				return fmt.Errorf("unable to convert parameters layer to ollama")
			}

			paramLayer := stream.NewLayer(
				out,
				stream.WithCompressionLevel(gzip.DefaultCompression),
				stream.WithMediaType(types.OLLAMA_MEDIA_TYPE_PARAMETERS),
			)

			l = paramLayer
			fallthrough
		default:
			sd, err := writeLayerBlob(blobsFolder, l, string(mt))
			if err != nil {
				return fmt.Errorf("unable to write layer blob: %w", err)
			}
			schemaLayers = append(schemaLayers, *sd)
		}
	}

	config := &types.OllamaConfig{
		ModelFormat:   "gguf",
		ModelFamilly:  "llama",
		ModelFamilies: []string{"llama"},
		ModelType:     "7B",
		FileType:      "Q4_0",
		Architecture:  "amd64",
		OS:            "linux",
		RootFS: types.RootFS{
			Type:    "layers",
			DiffIDs: []string{},
		},
	}

	for _, l := range layers {
		diff, err := l.DiffID()
		if err != nil {
			return fmt.Errorf("unable to get diff id from layer: %s", err)
		}

		config.RootFS.DiffIDs = append(config.RootFS.DiffIDs, diff.String())
	}

	size, err := config.Size()
	if err != nil {
		return fmt.Errorf("unable to generate size from config: %s", err)
	}

	dgst, err := config.Digest()
	if err != nil {
		return fmt.Errorf("unable to generate digest from config: %s", err)
	}

	// write the config
	err = config.WriteToDisk(blobsFolder)
	if err != nil {
		return fmt.Errorf("unable to write config: %s", err)
	}

	configDescriptor := manifest.Schema2Descriptor{
		MediaType: manifest.DockerV2Schema2ConfigMediaType,
		Size:      int64(size),
		Digest:    digest.Digest(dgst),
	}

	schema := &manifest.Schema2{
		SchemaVersion:     2,
		MediaType:         manifest.DockerV2Schema2MediaType,
		ConfigDescriptor:  configDescriptor,
		LayersDescriptors: schemaLayers,
	}

	tag := ref.Identifier()
	f, err := os.Create(path.Join(manifestFolder, tag))
	if err != nil {
		return fmt.Errorf("unable to open manefest file for writing: %s", err)
	}

	d := json.NewEncoder(f)

	err = d.Encode(schema)
	if err != nil {
		return fmt.Errorf("unable to encode manefest: %s", err)
	}

	return nil
}

// writes a layer as a blob and returns the schema descriptor
func writeLayerBlob(blobPath string, layer v1.Layer, layerType string) (*manifest.Schema2Descriptor, error) {
	// write the layer blob we need to do this first so that the digest and
	// can be computed, first we write to a temp file and then rename
	rc, err := layer.Compressed()
	if err != nil {
		return nil, fmt.Errorf("unable to get reader from layer: %w", err)
	}

	// the layer is compressed, get a gzip reader to decompress as we write it
	gzrc, err := gzip.NewReader(rc)
	if err != nil {
		return nil, fmt.Errorf("unable to create gzipped reader: %w", err)
	}

	// write to a temporary file as the digest is not available until the layer
	// has been read
	tempPath := path.Join(blobPath, "sha256-temp")
	os.Remove(tempPath)

	f, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open layer blob for writing: %s", err)
	}

	// write the blob
	io.Copy(f, gzrc)
	f.Close()
	rc.Close()

	// create the manifest
	sd := manifest.Schema2Descriptor{}

	switch layerType {
	case types.KAPSULE_MEDIA_TYPE_PARAMETERS:
		sd.MediaType = types.OLLAMA_MEDIA_TYPE_PARAMETERS
	case types.KAPSULE_MEDIA_TYPE_MODEL:
		sd.MediaType = types.OLLAMA_MEDIA_TYPE_MODEL
	case types.KAPSULE_MEDIA_TYPE_LICENCE:
		sd.MediaType = types.OLLAMA_MEDIA_TYPE_LICENCE
	case types.KAPSULE_MEDIA_TYPE_TEMPLATE:
		sd.MediaType = types.OLLAMA_MEDIA_TYPE_TEMPLATE
	default:
		sd.MediaType = layerType
	}

	d, err := layer.DiffID()
	if err != nil {
		return nil, fmt.Errorf("unable to get digest from layer: %w", err)
	}
	sd.Digest = digest.Digest(d.String())

	s, err := layer.Size()
	if err != nil {
		return nil, fmt.Errorf("unable to get size from layer: %w", err)
	}
	sd.Size = int64(s)

	// rename the blob now the digest is available
	lName := path.Join(blobPath, fmt.Sprintf("sha256-%s", sd.Digest.Encoded()))
	err = os.Rename(tempPath, lName)
	if err != nil {
		return nil, fmt.Errorf("unable to rename blob: %s", err)
	}

	return &sd, nil
}
