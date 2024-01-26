package configmanager

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	configModels "com.code.vidmicro/com.code.vidmicro/settings/configmanager/cofingModels"
	"com.code.vidmicro/com.code.vidmicro/settings/utilsdatatypes"
	"github.com/bytedance/sonic"
)

type config struct {
	Address                      string                      `json:"address"`
	Database                     configModels.DatabaseConfig `json:"database"`
	ClassName                    string
	Controllers                  []string                       `json:"controllers"`
	SessionKey                   string                         `json:"sessionKey"`
	RegisteredTopics             []string                       `json:"registeredTopics"`
	PublishingTopics             []string                       `json:"publishingTopics"`
	PrintWarning                 bool                           `json:"printWarning"`
	PrintInfo                    bool                           `json:"printInfo"`
	SubscribedTopics             map[string]interface{}         `json:"subscribedTopics"`
	MicroServiceName             string                         `json:"microServiceName"`
	PublisherBatchSize           int64                          `json:"publisherBatchSize"`
	NatsServerAddress            string                         `json:"natsServer"`
	IsProduction                 bool                           `json:"isProduction"`
	SessionSecret                string                         `json:"sessionSecret"`
	Redis                        configModels.RedisConnection   `json:"redis"`
	ServiceLogName               string                         `json:"serviceLogName"`
	MapApis                      map[string][]string            `json:"apis"`
	TokenExpiry                  int64                          `json:"tokenExpiry"`
	MapAcl                       map[string]map[string][]string `json:"acl"`
	S3Settings                   configModels.S3Connection      `json:"s3Settings"`
	PageSize                     int64                          `json:"pageSize"`
	TitlesPostfix                string                         `json:"titlesPostfix"`
	RedisSeprator                string                         `json:"redisSeprator"`
	SingleTitlePostfix           string                         `json:"singleTitlePostfix"`
	LanguagePostfix              string                         `json:"languagePostfix"`
	StatusPostfix                string                         `json:"statusPostfix"`
	TitleExpiryTime              int                            `json:"titleExpiryTime"`
	LanguageMetadataPostfix      string                         `json:"languageMetadataPostfix"`
	LanguageMetaExpiryTime       int                            `json:"languageMetaExpiryTime"`
	ContentTypePostfix           string                         `json:"contentTypePostfix"`
	ContentTitleLanguagesPostfix string                         `json:"contentTitleLanguagesPostfix"`
	TitlesLanguageExpirationTime int                            `json:"titlesLanguageExpirationTime"`
	ContentPostFix               string                         `json:"contentPostfix"`
	SessionExpirySeconds         int64                          `json:"sessionExpirySeconds"`
	EmailConfig                  configModels.EmailConfig       `json:"emailConfig"`
	EmailVerificationTokenExpiry int64                          `json:"emailVerificationTokenExpiry"`
	EmailVerificationURL         string                         `json:"emailVerificationURL"`
	EmailBody                    string                         `json:"emailBody"`
	EmailSubject                 string                         `json:"emailSubject"`
	ResetPasswordEmailBody       string                         `json:"resetPasswordEmailBody"`
	ResetPasswordEmailSubject    string                         `json:"resetPasswordEmailSubject"`
	TitlesContentPostfix         string                         `json:"titlesContentPostfix"`
	Meilisearch                  configModels.MeilisearchConfig `json:"meilisearch"`
	MeilisearchIndex             string                         `json:"meiliSearchIndex"`
	MaxMeiliSearchUpdates        int64                          `json:"maxMeiliSearchUpdates"`
	GenresPostfix                string                         `json:"genresPostfix"`
	TypePostfix                  string                         `json:"typePostfix"`
	TitlesMetaPostfix            string                         `json:"titlesMetaPostfix"`
	AllowedExtensions            map[string]bool                `json:"allowedExtensions"`
	AllowedSizeInMbs             int                            `json:"allowedSizeInMbs"`
	GoogleLoginConfig            configModels.GoogleLoginConfig `json:"googleLoginConfig"`
	Acl                          map[string]map[string]*utilsdatatypes.Set
	Apis                         map[string]*utilsdatatypes.Set
	PasswordTokenExpiry          int64
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

	decoder := sonic.ConfigDefault.NewDecoder(configFile)
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

	decoder = sonic.ConfigDefault.NewDecoder(globalConfigFile)
	err = decoder.Decode(&c)

	if err != nil {
		fmt.Println("Error decoding Global JSON:", err)
		return
	}

	c.ClassName = serviceName

	instance.Acl = make(map[string]map[string]*utilsdatatypes.Set)
	for k, v := range instance.MapAcl {
		rawSet := utilsdatatypes.NewSet()
		instance.Acl[k] = make(map[string]*utilsdatatypes.Set)
		for innerK, innerV := range v {
			for _, val := range innerV {
				rawSet.Add(val)
			}
			instance.Acl[k][innerK] = rawSet
		}
	}

	instance.Apis = make(map[string]*utilsdatatypes.Set)

	for k, v := range instance.MapApis {
		rawSet := utilsdatatypes.NewSet()
		for _, val := range v {
			rawSet.Add(val)
		}
		instance.Apis[k] = rawSet
	}
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
