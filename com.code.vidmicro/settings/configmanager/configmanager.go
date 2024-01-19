package configmanager

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	configModels "com.code.vidmicro/com.code.vidmicro/settings/configmanager/cofingModels"
)

type config struct {
	Server             string                      `json:"address"`
	Database           configModels.DatabaseConfig `json:"database"`
	ClassName          string
	Controllers        []string                     `json:"controllers"`
	SessionKey         string                       `json:"sessionKey"`
	RegisteredTopics   []string                     `json:"registeredTopics"`
	PublishingTopics   []string                     `json:"publishingTopics"`
	PrintWarning       bool                         `json:"printWarning"`
	PrintInfo          bool                         `json:"printInfo"`
	SubscribedTopics   map[string]interface{}       `json:"subscribedTopics"`
	MicroServiceName   string                       `json:"microServiceName"`
	PublisherBatchSize int64                        `json:"publisherBatchSize"`
	NatsServerAddress  string                       `json:"natsServer"`
	IsProduction       bool                         `json:"isProduction"`
	SessionSecret      string                       `json:"sessionSecret"`
	Redis              configModels.RedisConnection `json:"redis"`
	ServiceLogName     string                       `json:"serviceLogName"`
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

func (c *config) Setup() {
	name, path, serviceName := c.getConfigNameAndPath()

	configFile, err := os.Open(path + name + ".json")
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

	globoalConfigPath := "config/" + name + ".json"
	globalConfigFile, err := os.Open(globoalConfigPath)
	if err != nil {
		log.Println("Error opening global config file:", err)
		return
	}
	defer globalConfigFile.Close()

	decoder = json.NewDecoder(globalConfigFile)
	err = decoder.Decode(&c)

	if err != nil {
		fmt.Println("Error decoding Global JSON:", err)
		return
	}

	c.ClassName = serviceName
}

// GetConfigNameAndPath get the config name on the basis of flag
func (c *config) getConfigNameAndPath() (string, string, string) {
	serverType := flag.String("env", "DEV", "use development server by default")
	configPath := flag.String("con", "AUTHSERVICE", "use Uploader server by default")

	var conName string
	var conPath string
	flag.Parse()
	switch *serverType {
	case "DEV":
		conName = "dev"
	case "UAT":
		conName = "uat"
	case "PROD":
		conName = "prod"
	}
	switch *configPath {
	case "AUTHSERVICE":
		conPath = "com.code.vidmicro/services/authservice/config/"
	case "TITLESSERVICE":
		conPath = "com.code.vidmicro/services/titlesservice/config/"
	case "CONTENTSERVICE":
		conPath = "com.code.vidmicro/services/contentservice/config/"
	}

	return conName, conPath, *configPath
}
