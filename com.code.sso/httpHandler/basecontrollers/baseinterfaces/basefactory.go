package baseinterfaces

type BaseControllerFactory interface {
	GetController(string) (Controller, error)
}
