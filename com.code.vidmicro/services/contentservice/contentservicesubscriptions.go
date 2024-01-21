package contentservice

import (
	"encoding/json"
	"log"
	"reflect"

	"com.code.vidmicro/com.code.vidmicro/app/controllers"
	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"com.code.vidmicro/com.code.vidmicro/settings/topics"
	"github.com/nats-io/nats.go"
)

type ContentServiceSubscriptions struct {
}

func (ts ContentServiceSubscriptions) RegisterSubscriptions() error {
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

func (ts ContentServiceSubscriptions) HandleLanguageCreated(msg *nats.Msg) {

	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleLanguageCreated:", converted)
		langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		languageController := controller.(*controllers.LanguageController)
		languageController.Add(languageController.GetDBName(), languageController.GetCollectionName(), langData, false)
	}
}

func (ts ContentServiceSubscriptions) HandleLanguageUpdated(msg *nats.Msg) {
	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleLanguageUpdated:", converted)
		langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		languageController := controller.(*controllers.LanguageController)
		languageController.UpdateLanguage(langData)
	}
}

func (ts ContentServiceSubscriptions) HandleLanguageDeleted(msg *nats.Msg) {
	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleLanguageDeleted:", converted)
		langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		languageController := controller.(*controllers.LanguageController)
		languageController.DeleteLanguage(langData)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleCreated(msg *nats.Msg) {
	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleCreated:", converted)
		// langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		// controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		// languageController := controller.(*controllers.LanguageController)
		// languageController.DeleteLanguage(langData)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleUpdated(msg *nats.Msg) {
	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleUpdated:", converted)
		// langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		// controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		// languageController := controller.(*controllers.LanguageController)
		// languageController.DeleteLanguage(langData)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleDeleted(msg *nats.Msg) {
	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleDeleted:", converted)
		// langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		// controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		// languageController := controller.(*controllers.LanguageController)
		// languageController.DeleteLanguage(langData)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleLanguageDeleted(msg *nats.Msg) {
	var data []interface{}
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleLanguageDeleted:", converted)
		// langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		// controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		// languageController := controller.(*controllers.LanguageController)
		// languageController.DeleteLanguage(langData)
	}
}
