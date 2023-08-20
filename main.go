package main

import (
	"log"

	"com.code.sso/com.code.sso/config"
	"com.code.sso/com.code.sso/httpHandler"
)

func main() {
	log.Println("starting server")
	config.GetInstance().Setup("setup/prod.json")
	httpHandler.GetInstance().Start()
}
