// twitterauth.go

package twitterauth

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/mrjones/oauth"
)

var (
	instance *OAuth1aConfig
	once     sync.Once
)

type OAuth1aConfig struct {
	oauthConfigs       *oauth.Consumer
	TwitterUserInfoUrl string
	TwitterCallbackUrl string
}

func GetInstance() *OAuth1aConfig {
	once.Do(func() {
		config := configmanager.GetInstance().TwitterAuthConfig

		instance = &OAuth1aConfig{
			oauthConfigs: oauth.NewConsumer(
				config.TwitterConsumerKey,
				config.TwitterConsumerSecret,
				oauth.ServiceProvider{
					RequestTokenUrl:   config.TwitterRequestTokenUrl,
					AuthorizeTokenUrl: config.TwitterAuthorizeUrl,
					AccessTokenUrl:    config.TwitterAccessTokenUrl,
				},
			),
			TwitterUserInfoUrl: config.TwitterUserInfoUrl,
			TwitterCallbackUrl: config.TwitterCallbackUrl,
		}
	})
	return instance
}

func (o *OAuth1aConfig) Twitter() *oauth.Consumer {
	return o.oauthConfigs
}

func (o *OAuth1aConfig) GetTwitterUserInfoUrl() string {
	return o.TwitterUserInfoUrl
}

func (o *OAuth1aConfig) GetTwitterCallbackUrl() string {
	return o.TwitterCallbackUrl
}
