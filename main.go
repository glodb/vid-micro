package main

import (
	"log"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/servicehandler"
	"com.code.vidmicro/com.code.vidmicro/settings/topics"
)

func main() {

	log.Println("starting server")
	configmanager.GetInstance().Setup()

	topics.GetInstance().RegisterTopics()
	topics.GetInstance().RegisterPublishingTopics()

	base, err := servicehandler.GetInstance().InitializeService(configmanager.GetInstance().ClassName)
	// fmt.Println(err)
	if err != nil {
		//fmt.Println(config.GetString("initClassName"))
	} else {
		// fmt.Println("base run")
		base.Run()
	}
	base.Stop()
}
