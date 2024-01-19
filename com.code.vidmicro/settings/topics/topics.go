package topics

import (
	"log"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

type Topics struct {
	registeredTopics  map[string]bool
	publishableTopics map[string]bool
	listeningTopics   map[string]interface{}
}

var once sync.Once
var instance *Topics

//Singleton. Returns a single object of Topic
func GetInstance() *Topics {

	once.Do(func() {
		instance = &Topics{}
		instance.registeredTopics = make(map[string]bool)
		instance.publishableTopics = make(map[string]bool)
		instance.listeningTopics = make(map[string]interface{})
	})
	return instance
}

func (t *Topics) RegisterTopics() {
	registerdTopics := configmanager.GetInstance().RegisteredTopics

	if registerdTopics != nil {

		for i := range registerdTopics {
			t.registeredTopics[registerdTopics[i]] = true
		}

	} else {
		if configmanager.GetInstance().PrintWarning {
			log.Println("No Register Topics found")
		}
	}
}

func (t *Topics) GetSubscribedTopics() map[string]interface{} {
	subscribedTopics := configmanager.GetInstance().SubscribedTopics
	newSub := make(map[string]interface{})

	if subscribedTopics == nil {
		return newSub
	}

	for key, element := range subscribedTopics {
		newSub[key] = element
	}
	if subscribedTopics != nil {
		t.listeningTopics = newSub
	} else {
		return nil
	}
	return t.listeningTopics
}

func (t *Topics) RegisterPublishingTopics() {
	publishingTopics := configmanager.GetInstance().PublishingTopics
	if publishingTopics != nil {
		for i := range publishingTopics {
			t.publishableTopics[publishingTopics[i]] = true
		}
	} else {
		//logger.Info("No Register Topics found")
	}
}

func (t *Topics) ValidatePublishableTopics(key string) bool {
	if t.ValidateTopic(key) {
		if _, ok := t.publishableTopics[key]; ok {
			return true
		}
	}
	return false
}

func (t *Topics) ValidateTopic(key string) bool {
	if _, ok := t.registeredTopics[key]; ok {
		return true
	}
	return false
}
