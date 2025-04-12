package slackbot

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SlackAppID             string `envconfig:"SLACK_APP_ID" required:"true"`
	SlackClientID          string `envconfig:"SLACK_CLIENT_ID" required:"true"`
	SlackClientSecret      string `envconfig:"SLACK_CLIENT_SECRET" required:"true"`
	SlackSigninSecret      string `envconfig:"SLACK_SIGNIN_SECRET" required:"true"`
	SlackVerificationToken string `envconfig:"SLACK_VERIFICATION_TOKEN" required:"true"`
	SlackBotUserOAuthToken string `envconfig:"SLACK_BOT_USER_OAUTH_TOKEN" required:"true"`
}

func LoadConfig() (*Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
