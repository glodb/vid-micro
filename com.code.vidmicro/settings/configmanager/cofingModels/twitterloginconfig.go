package configModels

type TwitterLoginConfig struct {
	ClientId            string   `json:"clientId"`
	ClientSecret        string   `json:"clientSecret"`
	RedirectUrl         string   `json:"redirectUrl"`
	Scopes              []string `json:"scopes"`
	AuthUrl             string   `json:"authUrl"`
	TokenUrl            string   `json:"tokenUrl"`
	CodeChallengeMethod string   `json:"codeChallengeMethod"`
	TwitterUserInfoUrl  string   `json:"twitterUserInfoUrl"`
}
