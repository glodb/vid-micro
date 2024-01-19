package serviceutils

import (
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

type ServiceUtils struct {
	nat            *nats.Conn
	eventPublisher EventPublisher
	exitChannel    chan bool
}

var instance *ServiceUtils
var once sync.Once

func GetInstance() *ServiceUtils {
	once.Do(func() {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			log.Fatal("Unable to connect to nats")
		}
		instance = &ServiceUtils{}
		instance.nat = nc
		instance.eventPublisher = EventPublisher{}
		instance.eventPublisher.New()
	})
	return instance
}

func (s *ServiceUtils) GetNat() *nats.Conn {
	return s.nat
}

func (s *ServiceUtils) serviceInitialization() {

}

func (s *ServiceUtils) Shutdown() {
	close(s.exitChannel)
}

func (s *ServiceUtils) RunService() {
}
