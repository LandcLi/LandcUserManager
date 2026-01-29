package client

import (
	"github.com/LandcLi/LandcGoTools/mail"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	*mail.Config `yaml:",inline"`
	From         string `yaml:"from"`
}

var mailClient *mail.Client

func InitMailClientByConfig(config *mail.Config) {
	mailClient = mail.NewClientWithConfig(config)
}
func InitMailClientBySet(client *mail.Client) {
	mailClient = client
}
func GetMailClient() *mail.Client {
	return mailClient
}
