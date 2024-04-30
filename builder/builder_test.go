package builder

import (
	"compress/gzip"
	"io"
	"os"
	"path"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/nicholasjackson/kapsule/modelfile"
	pm "github.com/nicholasjackson/kapsule/modelfile/mocks"
	kt "github.com/nicholasjackson/kapsule/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupBuilder(t *testing.T) (builder Builder, mockParser *pm.Parser, context string, output string) {
	td := t.TempDir()
	o := path.Join(td, "output")
	ctx := path.Join(td, "context")

	os.MkdirAll(o, os.ModePerm)
	os.MkdirAll(ctx, os.ModePerm)

	// create an example model
	mf := path.Join(ctx, "model.gguf")
	os.WriteFile(mf, []byte("blah"), os.ModePerm)

	model := &modelfile.ModelFile{
		From:       "./model.gguf",
		Template:   "[Inst] Something [/Inst]",
		Parameters: map[string][]string{"a": []string{"1"}, "b": []string{"2"}},
	}

	mp := &pm.Parser{}
	mp.On("Parse", mock.Anything).Return(model, nil)

	b := &BuilderImpl{mp}

	return b, mp, ctx, o
}

func TestBuildLoadsModelFile(t *testing.T) {
	b, mb, _, _ := setupBuilder(t)

	_, err := b.Build("./blah.modelfile", t.TempDir(), t.TempDir())
	require.NoError(t, err)

	mb.AssertCalled(t, "Parse", "./blah.modelfile")
}

func TestBuildAddsModelLayer(t *testing.T) {
	b, _, ctx, output := setupBuilder(t)

	img, err := b.Build("./blah.modelfile", ctx, output)
	require.NoError(t, err)
	require.NotNil(t, img)

	fl, _ := img.Layers()

	mt, _ := fl[0].MediaType()
	require.Equal(t, types.MediaType(kt.KAPSULE_MEDIA_TYPE_MODEL), mt)

	rc, err := fl[0].Compressed()
	require.NoError(t, err)

	// uncompress the reader
	gzr, err := gzip.NewReader(rc)
	require.NoError(t, err)

	d, err := io.ReadAll(gzr)
	require.NoError(t, err)
	require.Contains(t, string(d), "blah")
}

func TestBuildAddsTemplateLayer(t *testing.T) {
	b, _, ctx, output := setupBuilder(t)

	img, err := b.Build("./blah.modelfile", ctx, output)
	require.NoError(t, err)
	require.NotNil(t, img)

	fl, _ := img.Layers()

	mt, _ := fl[1].MediaType()
	require.Equal(t, types.MediaType(kt.KAPSULE_MEDIA_TYPE_TEMPLATE), mt)

	rc, err := fl[1].Compressed()
	require.NoError(t, err)

	// uncompress the reader
	gzr, err := gzip.NewReader(rc)
	require.NoError(t, err)

	d, err := io.ReadAll(gzr)
	require.NoError(t, err)
	require.Equal(t, "[Inst] Something [/Inst]", string(d))
}

func TestBuildAddsParametersLayer(t *testing.T) {
	b, _, ctx, output := setupBuilder(t)

	img, err := b.Build("./blah.modelfile", ctx, output)
	require.NoError(t, err)
	require.NotNil(t, img)

	fl, _ := img.Layers()

	mt, _ := fl[2].MediaType()
	require.Equal(t, types.MediaType(kt.KAPSULE_MEDIA_TYPE_PARAMETERS), mt)

	rc, err := fl[2].Compressed()
	require.NoError(t, err)

	// uncompress the reader
	gzr, err := gzip.NewReader(rc)
	require.NoError(t, err)

	d, err := io.ReadAll(gzr)
	require.NoError(t, err)
	require.JSONEq(t, string(d), `{"a": ["1"], "b": ["2"]}`)
}
