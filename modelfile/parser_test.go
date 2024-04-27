package modelfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsesValidModelFile(t *testing.T) {
	p := &ParserImpl{}

	_, err := p.Parse("../test_fixtures/modelfile/basic_with_template.modelfile")
	require.NoError(t, err)
}
