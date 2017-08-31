package notifier

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"

	"github.com/giantswarm/draughtsman-operator/flag"
	slacknotifier "github.com/giantswarm/draughtsman-operator/service/notifier/slack"
	"github.com/giantswarm/draughtsman-operator/service/notifier/spec"
)

// Config represents the configuration used to create a Notifier.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

	Type spec.NotifierType
}

// DefaultConfig provides a default configuration to create a new Notifier
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Settings.
		Flag:  nil,
		Viper: nil,
	}
}

// New creates a new configured Notifier.
func New(config Config) (spec.Notifier, error) {
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "viper must not be empty")
	}

	var err error

	var newNotifier spec.Notifier
	switch config.Type {
	case slacknotifier.SlackNotifierType:
		slackConfig := slacknotifier.DefaultConfig()

		slackConfig.Logger = config.Logger
		slackConfig.SlackClient = slack.New(config.Viper.GetString(config.Flag.Service.Notifier.Slack.Token))

		slackConfig.Channel = config.Viper.GetString(config.Flag.Service.Notifier.Slack.Channel)
		slackConfig.Emoji = config.Viper.GetString(config.Flag.Service.Notifier.Slack.Emoji)
		// TODO configure env in separate PR when adding eventer to sort out deployment event status updates.
		//slackConfig.Environment = config.Viper.GetString(config.Flag.Service.Eventer.Environment)
		slackConfig.Environment = "TODO"
		slackConfig.Username = config.Viper.GetString(config.Flag.Service.Notifier.Slack.Username)

		newNotifier, err = slacknotifier.New(slackConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	default:
		return nil, microerror.Maskf(invalidConfigError, "notifier type not implemented")
	}

	return newNotifier, nil
}
