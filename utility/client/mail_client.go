package client

import (
	"github.com/LandcLi/LandcGoTools/mail"
)

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
