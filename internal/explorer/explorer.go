package explorer

type Explorer interface {
	Emit() error
	TargetEmit(target string) error
	Stop()
}
