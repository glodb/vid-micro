package configModels

type TwitterAuthConfig struct {
	TwitterCallbackUrl     string `json:"twitterCallbackUrl"`
	TwitterConsumerKey     string `json:"twitterConsumerKey"`
	TwitterConsumerSecret  string `json:"twitterConsumerSecret"`
	TwitterRequestTokenUrl string `json:"twitterRequestTokenUrl"`
	TwitterAuthorizeUrl    string `json:"twitterAuthorizeURL"`
	TwitterAccessTokenUrl  string `json:"twitterAccessTokenUrl"`
	TwitterUserInfoUrl     string `json:"twitterUserInfoUrl"`
}
