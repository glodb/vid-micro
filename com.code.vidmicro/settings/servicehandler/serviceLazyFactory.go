package servicehandler

import (
	"errors"
	"strings"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/services/authservice"
)

type services struct {
	authservice authservice.AuthService
}

var instance *services
var once sync.Once

//Singleton. Returns a single object of Factory
func GetInstance() *services {

	once.Do(func() {
		instance = &services{}
	})
	return instance
}

func (c *services) InitializeService(serviceType string) (ServiceBase, error) {

	switch strings.ToUpper(serviceType) {
	case "AUTHSERVICE":
		if c.authservice == (authservice.AuthService{}) {
			c.authservice = authservice.AuthService{}
		}
		return &c.authservice, nil
	}

	return nil, errors.New("Not known service found")
}
