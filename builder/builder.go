package builder

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/nicholasjackson/kapsule/encryptor"
	"github.com/nicholasjackson/kapsule/modelfile"
	"github.com/nicholasjackson/kapsule/types"
)

// Builder defines an interface for generating OCI images
//
//go:generate mockery --name Builder
type Builder interface {
	// Build an image using the given modelfile path and write to the output path
	Build(model, context string) (v1.Image, error)
}

// BuilderImpl is a concrete implementation of the Builder interface
type BuilderImpl struct {
	parser    modelfile.Parser
	encryptor encryptor.EncryptorImpl
}

func NewBuilder() Builder {
	return &BuilderImpl{
		&modelfile.ParserImpl{},
		encryptor.EncryptorImpl{},
	}
}

func (b *BuilderImpl) Build(model, context string) (v1.Image, error) {
	base := empty.Image

	// parse the modelfile
	mf, err := b.parser.Parse(model)
	if err != nil {
		return nil, fmt.Errorf("unable to load modelfile: %s", err)
	}

	// add the model in FROM
	fPath := path.Join(context, mf.From)
	f, err := os.Open(fPath)
	if err != nil {
		return nil, fmt.Errorf("unable to find file: %s defined in FROM: %s", mf.From, err)
	}

	fromLayer := stream.NewLayer(
		f,
		stream.WithCompressionLevel(gzip.DefaultCompression),
		stream.WithMediaType(types.KAPSULE_MEDIA_TYPE_MODEL),
	)

	image, err := mutate.AppendLayers(base, fromLayer)
	if err != nil {
		return nil, fmt.Errorf("unable add FROM layer: %s", err)
	}

	if mf.Template != "" {
		templateLayer := stream.NewLayer(
			io.NopCloser(bytes.NewReader([]byte(mf.Template))),
			stream.WithCompressionLevel(gzip.DefaultCompression),
			stream.WithMediaType(types.KAPSULE_MEDIA_TYPE_TEMPLATE),
		)

		image, err = mutate.AppendLayers(image, templateLayer)
		if err != nil {
			return nil, fmt.Errorf("unable add TEMPLATE layer: %s", err)
		}
	}

	if len(mf.Parameters) > 0 {
		jp, err := json.Marshal(mf.Parameters)
		if err != nil {
			return nil, fmt.Errorf("unable to add PARAMETERS layer: %s", err)
		}

		paramsLayer := stream.NewLayer(
			io.NopCloser(bytes.NewReader(jp)),
			stream.WithCompressionLevel(1),
			stream.WithMediaType(types.KAPSULE_MEDIA_TYPE_PARAMETERS),
		)

		image, err = mutate.AppendLayers(image, paramsLayer)
		if err != nil {
			return nil, fmt.Errorf("unable add PARAMETERS layer: %s", err)
		}
	}

	return image, nil
}
