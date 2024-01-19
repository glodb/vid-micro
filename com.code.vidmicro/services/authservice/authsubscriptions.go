package authservice

import (
	"log"
	"reflect"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"com.code.vidmicro/com.code.vidmicro/settings/topics"
	"github.com/nats-io/nats.go"
)

type AuthSubscriptions struct {
}

func (ts AuthSubscriptions) RegisterSubscriptions() error {
	subTopics := topics.GetInstance().GetSubscribedTopics()
	for k, v := range subTopics {
		log.Println(k, v)
		if topics.GetInstance().ValidateTopic(k) {
			m := reflect.ValueOf(&ts).MethodByName(v.(string))
			mCallable := m.Interface().(func(msg *nats.Msg))
			serviceutils.GetInstance().GetNat().QueueSubscribe(k, configmanager.GetInstance().ClassName, mCallable)
		} else {
			log.Println("This Topic is not registered")
		}

	}
	return nil
}

func (ts AuthSubscriptions) HandleAuthData(msg *nats.Msg) {
	log.Println("Subscribed data", string(msg.Data))
}
