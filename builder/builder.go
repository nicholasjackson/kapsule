package builder

import (
	"fmt"
	"os"
	"path"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/nicholasjackson/kapsule/modelfile"
)

const MEDIA_TYPE_MODEL = "application/vnd.kapsule.image.model"

// Builder defines an interface for generating OCI images
//
//go:generate mockery --name Builder
type Builder interface {
	// Build an image using the given modelfile path and write to the output path
	Build(model, context, output string) (v1.Image, error)
}

// BuilderImpl is a concrete implementation of the Builder interface
type BuilderImpl struct {
	parser modelfile.Parser
}

func NewBuilder() Builder {
	return &BuilderImpl{&modelfile.ParserImpl{}}
}

func (b *BuilderImpl) Build(model, context, output string) (v1.Image, error) {
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

	fromLayer := stream.NewLayer(f, stream.WithCompressionLevel(1), stream.WithMediaType(MEDIA_TYPE_MODEL))

	image, err := mutate.AppendLayers(base, fromLayer)
	if err != nil {
		return nil, fmt.Errorf("unable add FROM layer: %s", err)
	}

	return image, nil
}
