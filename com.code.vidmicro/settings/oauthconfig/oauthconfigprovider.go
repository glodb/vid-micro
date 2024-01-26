package oauthconfig

import (
	"errors"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/oauthconfig/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	instance *OAuthConfig
	once     sync.Once
)

type OAuthConfig struct {
	oauthConfigs map[services.ServiceType]*oauth2.Config
}

func GetInstance() *OAuthConfig {
	once.Do(func() {
		instance = &OAuthConfig{}
		instance.oauthConfigs = make(map[services.ServiceType]*oauth2.Config)
	})
	return instance
}

func (oauth *OAuthConfig) GetOAuth2Config(serviceType services.ServiceType) (*oauth2.Config, error) {

	if config, ok := oauth.oauthConfigs[serviceType]; ok {
		return config, nil
	}

	var config *oauth2.Config
	switch serviceType {
	case services.Google:
		config = &oauth2.Config{
			ClientID:     configmanager.GetInstance().GoogleLoginConfig.ClientId,
			ClientSecret: configmanager.GetInstance().GoogleLoginConfig.ClientSecret,
			RedirectURL:  configmanager.GetInstance().GoogleLoginConfig.RedirectUrl,
			Scopes:       configmanager.GetInstance().GoogleLoginConfig.Scopes,
			Endpoint:     google.Endpoint,
		}
		oauth.oauthConfigs[serviceType] = config
		return config, nil
	default:
		return nil, errors.New("service is not registered")
	}
}
