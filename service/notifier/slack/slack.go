package slack

import (
	"fmt"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/nlopes/slack"

	"github.com/giantswarm/draughtsman-operator/service/notifier/spec"
)

const (
	// goodColour is the colour to use for success Slack messages.
	goodColour = "good"
	// dangerColour is the colour to use for failure Slack messages.
	dangerColour = "danger"
)

const (
	// failedMessageFormat is the format for failure Slack messages. Templated
	// with the error message itself.
	failedMessageFormat = "Encountered an error ```%v```"
	// successMessage is the message for success Slack messages.
	successMessage = "Successfully deployed"
	// titleFormat is the format for titles for Slack messages. Templated with the
	// repository name, and sha, e.g: "api - 12345".
	titleFormat = "%v - %v"
)

// SlackNotifierType is an Notifier that uses Slack.
var SlackNotifierType spec.NotifierType = "SlackNotifier"

// Config represents the configuration used to create a Slack Notifier..
type Config struct {
	// Dependencies.
	Logger      micrologger.Logger
	SlackClient Client

	// Settings.
	Channel     string
	Emoji       string
	Environment string
	Username    string
}

// DefaultConfig provides a default configuration to create a new Slack
// Notifier by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:      nil,
		SlackClient: nil,

		// Settings.
		Channel:     "",
		Emoji:       "",
		Environment: "",
		Username:    "",
	}
}

// New creates a new configured Slack Notifier.
func New(config Config) (*Notifier, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.SlackClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.SlackClient must not be empty")
	}

	// Settings.
	if config.Channel == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Channel must not be empty")
	}
	if config.Emoji == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Emoji must not be empty")
	}
	if config.Environment == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Environment must not be empty")
	}
	if config.Username == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Username must not be empty")
	}

	// TODO add common health check for this.
	config.Logger.Log("debug", "checking connection to Slack")
	_, err := config.SlackClient.AuthTest()
	if err != nil {
		return nil, microerror.Maskf(err, "could not authenticate with slack")
	}

	notifier := &Notifier{
		// Dependencies.
		logger:      config.Logger,
		slackClient: config.SlackClient,

		// Settings.
		channel:     config.Channel,
		emoji:       config.Emoji,
		environment: config.Environment,
		username:    config.Username,
	}

	return notifier, nil
}

// Notifier is an implementation of the Notifier interface,
// that uses Slack.
type Notifier struct {
	// Dependencies.
	logger      micrologger.Logger
	slackClient Client

	// Settings.
	channel     string
	emoji       string
	environment string
	username    string
}

func (n *Notifier) Failed(project spec.Project, errorMessage string) error {
	n.logger.Log("debug", "sending failed message to slack")
	return n.postSlackMessage(project, errorMessage)
}

func (n *Notifier) Success(project spec.Project) error {
	n.logger.Log("debug", "sending success message to slack")
	return n.postSlackMessage(project, "")
}

// postSlackMessage takes a DeploymentEvent and a possible error message,
// and posts a helpful message to the configured Slack channel.
func (n *Notifier) postSlackMessage(project spec.Project, errorMessage string) error {
	startTime := time.Now()
	defer updateSlackMetrics(startTime)

	success := false
	if errorMessage == "" {
		success = true
	}

	attachment := slack.Attachment{}

	attachment.Color = dangerColour
	if success {
		attachment.Color = goodColour
	}

	attachment.MarkdownIn = []string{"text"}

	attachment.Title = fmt.Sprintf(titleFormat, project.Name, project.Ref)
	attachment.Text = fmt.Sprintf(failedMessageFormat, errorMessage)
	if success {
		attachment.Text = successMessage
	}
	attachment.Footer = n.environment

	params := slack.PostMessageParameters{}

	params.Username = n.username
	params.IconEmoji = n.emoji
	params.Attachments = []slack.Attachment{attachment}

	_, _, err := n.slackClient.PostMessage(n.channel, "", params)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
