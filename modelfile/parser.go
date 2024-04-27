package modelfile

import (
	"fmt"
	"os"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
)

type ModelFile struct {
	From      string
	Template  string
	Parameter map[string]string
}

//go:generate mockery --name Parser
type Parser interface {
	Parse(file string) (*ModelFile, error)
}

type ParserImpl struct{}

func (p *ParserImpl) Parse(file string) (*ModelFile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open modelfile: %w", err)
	}

	defer f.Close()

	r, err := parser.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse modelfile")
	}

	mf := &ModelFile{
		Parameter: map[string]string{},
	}

	s := shell.NewLex('\\')

	for _, c := range r.AST.Children {
		switch c.Value {
		case "FROM":
			mf.From = c.Next.Value
		case "TEMPLATE":
			w, _ := s.ProcessWords(c.Original, []string{})
			mf.Template = w[0]
		case "PARAMETER":
			w, _ := s.ProcessWords(c.Original, []string{})
			mf.Parameter["a"] = w[0]
		}
	}

	return nil, nil
}
