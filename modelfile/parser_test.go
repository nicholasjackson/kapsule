package modelfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsesFROMInModelFile(t *testing.T) {
	p := &ParserImpl{}

	m, err := p.Parse("../test_fixtures/modelfile/basic_with_template.modelfile")
	require.NoError(t, err)

	require.Equal(t, `./model.gguf`, m.From)
}

func TestModelfileWithBadFromReturnsError(t *testing.T) {
	p := &ParserImpl{}

	_, err := p.Parse("../test_fixtures/modelfile/basic_with_bad_from.modelfile")
	require.Error(t, err)
}

func TestParsesTemplateInModelFile(t *testing.T) {
	p := &ParserImpl{}

	m, err := p.Parse("../test_fixtures/modelfile/basic_with_template.modelfile")
	require.NoError(t, err)

	require.Equal(t, `[INST] {{ .System }} {{ .Prompt }} [/INST]`, m.Template)
}

func TestModelfileWithBadTemplateReturnsError(t *testing.T) {
	p := &ParserImpl{}

	_, err := p.Parse("../test_fixtures/modelfile/basic_with_bad_template.modelfile")
	require.Error(t, err)
}

func TestParsesParametersInModelFile(t *testing.T) {
	p := &ParserImpl{}

	m, err := p.Parse("../test_fixtures/modelfile/basic_with_template.modelfile")
	require.NoError(t, err)

	require.Len(t, m.Parameters["stop"], 2)
	require.Equal(t, `[/INST]`, m.Parameters["stop"][0])
	require.Equal(t, `[INST]`, m.Parameters["stop"][1])
}

func TestModelfileWithBadParametersReturnsError(t *testing.T) {
	p := &ParserImpl{}

	_, err := p.Parse("../test_fixtures/modelfile/basic_with_bad_parameters.modelfile")
	require.Error(t, err)
}
