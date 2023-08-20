package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	configModels "com.code.sso/com.code.sso/config/cofingModels"
)

type config struct {
	Server      configModels.ServerConfig   `json:"server"`
	Database    configModels.DatabaseConfig `json:"database"`
	Controllers []string                    `json:"controllers"`
	SessionKey  string                      `json:"sessionKey"`
}

var (
	instance *config
	once     sync.Once
)

func GetInstance() *config {
	// var instance
	once.Do(func() {
		instance = &config{}
	})
	return instance
}

func (c *config) Setup(path string) {
	configFile, err := os.Open(path)
	if err != nil {
		log.Println("Error opening config file:", err)
		return
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&c)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
}
