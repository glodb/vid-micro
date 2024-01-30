package servicehandler

import (
	"errors"
	"strings"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/services/authservice"
	"com.code.vidmicro/com.code.vidmicro/services/contentservice"
	"com.code.vidmicro/com.code.vidmicro/services/sessionscleanerservice"
	titlesservice "com.code.vidmicro/com.code.vidmicro/services/titlesservice"
)

type services struct {
	authservice           authservice.AuthService
	titlesservice         titlesservice.TitlesService
	contentserivce        contentservice.ContentService
	sessioncleanerservice sessionscleanerservice.SessionsCleanerService
}

var instance *services
var once sync.Once

// Singleton. Returns a single object of Factory
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
	case "TITLESSERVICE":
		if c.titlesservice == (titlesservice.TitlesService{}) {
			c.titlesservice = titlesservice.TitlesService{}
		}
		return &c.titlesservice, nil
	case "CONTENTSERVICE":
		if c.contentserivce == (contentservice.ContentService{}) {
			c.contentserivce = contentservice.ContentService{}
		}
		return &c.contentserivce, nil
	case "SESSIONSCLEANERSERVICE":
		if c.sessioncleanerservice == (sessionscleanerservice.SessionsCleanerService{}) {
			c.sessioncleanerservice = sessionscleanerservice.SessionsCleanerService{}
		}
		return &c.sessioncleanerservice, nil
	}

	return nil, errors.New("not known service found")
}
