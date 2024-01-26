package googleloginconfig

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	onceGoogleConfig sync.Once
	instance         *oauth2.Config
)

// GoogleConfig represents the Google API configuration.
// type GoogleConfig struct {
// 	ClientID     string
// 	ClientSecret string
// 	RedirectURL  string
// 	Scopes       []string
// 	Endpoint     string
// }

// GetGoogleConfig returns the Google API configuration.

func GetInstance() *oauth2.Config {
	onceGoogleConfig.Do(func() {
		instance = &oauth2.Config{
			ClientID:     configmanager.GetInstance().GoogleLoginConfig.ClientId,
			ClientSecret: configmanager.GetInstance().GoogleLoginConfig.ClientSecret,
			RedirectURL:  configmanager.GetInstance().GoogleLoginConfig.RedirectUrl,
			Scopes:       configmanager.GetInstance().GoogleLoginConfig.Scopes,
			Endpoint:     google.Endpoint,
		}
	})
	return instance
}
