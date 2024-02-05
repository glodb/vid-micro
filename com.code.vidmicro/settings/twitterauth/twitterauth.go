package twitterauth

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	configModels "com.code.vidmicro/com.code.vidmicro/settings/configmanager/cofingModels"
)

var (
	instance *configModels.TwitterAuthConfig
	once     sync.Once
)

func GetInstance() *configModels.TwitterAuthConfig {
	once.Do(func() {
		instance = &configModels.TwitterAuthConfig{
			TwitterConsumerKey:     configmanager.GetInstance().TwitterAuthConfig.TwitterConsumerKey,
			TwitterConsumerSecret:  configmanager.GetInstance().TwitterAuthConfig.TwitterConsumerSecret,
			TwitterRequestTokenUrl: configmanager.GetInstance().TwitterAuthConfig.TwitterRequestTokenUrl,
			TwitterAuthorizeUrl:    configmanager.GetInstance().TwitterAuthConfig.TwitterAuthorizeUrl,
			TwitterAccessTokenUrl:  configmanager.GetInstance().TwitterAuthConfig.TwitterAccessTokenUrl,
			TwitterUserInfoUrl:     configmanager.GetInstance().TwitterAuthConfig.TwitterUserInfoUrl,
			TwitterCallbackUrl:     configmanager.GetInstance().TwitterAuthConfig.TwitterCallbackUrl,
		}
	})
	return instance
}

func Twitter() *configModels.TwitterAuthConfig {
	once.Do(func() {

		instance = &configModels.TwitterAuthConfig{
			TwitterConsumerKey:     configmanager.GetInstance().TwitterAuthConfig.TwitterConsumerKey,
			TwitterConsumerSecret:  configmanager.GetInstance().TwitterAuthConfig.TwitterConsumerSecret,
			TwitterRequestTokenUrl: configmanager.GetInstance().TwitterAuthConfig.TwitterRequestTokenUrl,
			TwitterAuthorizeUrl:    configmanager.GetInstance().TwitterAuthConfig.TwitterAuthorizeUrl,
			TwitterAccessTokenUrl:  configmanager.GetInstance().TwitterAuthConfig.TwitterAccessTokenUrl,
			TwitterUserInfoUrl:     configmanager.GetInstance().TwitterAuthConfig.TwitterUserInfoUrl,
			TwitterCallbackUrl:     configmanager.GetInstance().TwitterAuthConfig.TwitterCallbackUrl,
		}
	})
	return instance
}
