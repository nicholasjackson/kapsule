package modelfile

import (
	"fmt"
	"os"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
)

type ModelFile struct {
	From       string
	Template   string
	Parameters map[string][]string
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
		Parameters: map[string][]string{},
	}

	s := shell.NewLex('\\')

	for _, c := range r.AST.Children {
		switch c.Value {
		case "FROM":
			w, _ := s.ProcessWords(c.Original, []string{})

			if len(w) != 2 {
				return nil, fmt.Errorf("FROM should be specified as FROM <path to model>")
			}

			mf.From = w[1]
		case "TEMPLATE":
			w, _ := s.ProcessWords(c.Original, []string{})

			if len(w) != 2 {
				return nil, fmt.Errorf("TEMPLATE should be specified as TEMPLATE \"The template to use for the model\"")
			}

			mf.Template = w[1]
		case "PARAMETER":
			w, _ := s.ProcessWords(c.Original, []string{})

			if len(w) != 3 {
				return nil, fmt.Errorf("PARAMETER should be specified as PRAMETER <key> <value>")
			}

			mf.Parameters[w[1]] = append(mf.Parameters[w[1]], w[2])
		}
	}

	return mf, nil
}
