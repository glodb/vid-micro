package main

import (
	"log"

	"com.code.sso/com.code.sso/httpHandler"
)

func main() {
	log.Println("starting server")
	httpHandler.GetInstance().Start()
}
