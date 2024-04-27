package builder

//go:generate mockery --name Builder
type Builder interface {
	Build(model string)
}
