package servicehandler

type ServiceBase interface {
	Run() error
	AssignSubscriber() error
	Stop()
}
