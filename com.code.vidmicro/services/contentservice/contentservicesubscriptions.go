package contentservice

import (
	"fmt"
	"log"
	"reflect"

	"com.code.vidmicro/com.code.vidmicro/app/controllers"
	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/serviceutils"
	"com.code.vidmicro/com.code.vidmicro/settings/topics"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/bytedance/sonic"
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
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleLanguageCreated:", converted)
		langData := models.LanguageContent{Id: int(converted["id"].(float64)), Name: converted["name"].(string), Code: converted["code"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.LanguageContent)
		languageController := controller.(*controllers.LanguageContentController)
		languageController.Add(languageController.GetDBName(), languageController.GetCollectionName(), langData, false)
		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", langData.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix), langData.EncodeRedisData())
	}
}

func (ts ContentServiceSubscriptions) HandleLanguageUpdated(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleLanguageUpdated:", converted)
		langData := models.LanguageContent{Id: int(converted["id"].(float64)), Name: converted["name"].(string), Code: converted["code"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.LanguageContent)
		languageController := controller.(*controllers.LanguageContentController)
		languageController.UpdateLanguage(langData)
		cache.GetInstance().Set(fmt.Sprintf("%d%s%s", langData.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix), langData.EncodeRedisData())
	}
}

func (ts ContentServiceSubscriptions) HandleLanguageDeleted(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleLanguageDeleted:", converted)
		langData := models.LanguageContent{Id: int(converted["id"].(float64)), Name: converted["name"].(string), Code: converted["code"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.LanguageContent)
		languageController := controller.(*controllers.LanguageContentController)
		languageController.DeleteLanguage(langData)

		keys := cache.GetInstance().GetKeys(fmt.Sprintf("*%s", configmanager.GetInstance().ContentPostFix))
		cache.GetInstance().DelMany(keys)

		cache.GetInstance().Del(fmt.Sprintf("%d%s%s", langData.Id, configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().LanguagePostfix))
	}
}

func (ts ContentServiceSubscriptions) HandleTitleCreated(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleCreated:", converted)
		titleData := models.TitlesSummary{Id: int(converted["Id"].(float64)), OriginalTitle: converted["original_title"].(string), Languages: utils.InterfaceArrayToIntArray(converted["languages"].([]interface{}))}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.TitlesSummary)
		_, err := controller.Add(controller.GetDBName(), controller.GetCollectionName(), titleData, false)
		log.Println(err)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleUpdated(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleUpdated:", converted)
		titleData := models.TitlesSummary{Id: int(converted["Id"].(float64)), OriginalTitle: converted["original_title"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.TitlesSummary)
		query := "UPDATE " + string(controller.GetCollectionName()) + " SET original_title=$1 where id=$2"
		controller.UpdateOne(controller.GetDBName(), controller.GetCollectionName(), string(query), []interface{}{titleData.OriginalTitle, titleData.Id}, false)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleDeleted(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleDeleted:", converted)

		titleData := models.TitlesSummary{Id: int(converted["Id"].(float64)), OriginalTitle: converted["original_title"].(string)}
		controller, _ := basecontrollers.GetInstance().GetController(baseconst.TitlesSummary)
		condition := map[string]interface{}{"id": titleData.Id}
		controller.DeleteOne(controller.GetDBName(), controller.GetCollectionName(), condition, false, false)

		contentController, _ := basecontrollers.GetInstance().GetController(baseconst.Content)
		condition = map[string]interface{}{"associated_title": titleData.Id}
		contentController.DeleteOne(contentController.GetDBName(), contentController.GetCollectionName(), condition, false, false)

		keys := cache.GetInstance().GetKeys(fmt.Sprintf("*%s%s", configmanager.GetInstance().RedisSeprator, configmanager.GetInstance().ContentPostFix))
		cache.GetInstance().DelMany(keys)
	}
}

func (ts ContentServiceSubscriptions) HandleTitleLanguageDeleted(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
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

func (ts ContentServiceSubscriptions) HandleTitleLanguageAdded(msg *nats.Msg) {
	var data []interface{}
	if err := sonic.Unmarshal(msg.Data, &data); err != nil {
		return
	}
	for _, s := range data {
		converted := s.(map[string]interface{})
		log.Println("HandleTitleLanguageAdded:", converted)
		// langData := models.LanguageMeta{LanguageId: int(converted["language_id"].(float64)), TitlesId: int(converted["titles_id"].(float64))}
		// controller, _ := basecontrollers.GetInstance().GetController(baseconst.Content)
		// titlesSummary := controller.(*controllers.TitlesSummaryController)

		// langData := models.Language{Id: converted["id"].(string), Name: converted["name"].(string), Code: converted["code"].(string)}
		// controller, _ := basecontrollers.GetInstance().GetController(baseconst.Language)
		// languageController := controller.(*controllers.LanguageController)
		// languageController.DeleteLanguage(langData)
	}
}
