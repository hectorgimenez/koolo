package runtype

type Run interface {
	Name() string
	Run() error
}
