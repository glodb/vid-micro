package configModels

type EmailConfig struct {
	SMTPServer  string `json:"smtpServer"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"fromAddress"`
	FromName    string `json:"fromName"`
	IsTLS       bool   `json:"isTLS"`
}
