package kapsule

type Image struct {
	Modelfile string
	Context   string
	Name      string
}

func (i *Image) Build(output string) error {
	return nil
}
