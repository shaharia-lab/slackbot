// Package slackbot provides a simple Slack bot implementation
// that can respond to messages and mentions in Slack channels.
// also reply to threads where it is mentioned to continue the conversation.
package slackbot

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// SlackAPI defines the interface for Slack client operations
type SlackAPI interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	AuthTest() (*slack.AuthTestResponse, error)
}

// Bot represents a Slack bot instance
type Bot struct {
	client        SlackAPI
	botUserID     string
	config        *Config
	activeThreads map[string]bool
	// Track processed events to avoid duplicates
	processedEvents      map[string]bool
	processedEventsMutex sync.Mutex
	log                  *slog.Logger
}

// NewBot creates a new Slack bot instance
func NewBot(config *Config, log *slog.Logger) (*Bot, error) {
	client := slack.New(config.SlackBotUserOAuthToken)
	return NewBotWithClient(client, config, log)
}

// NewBotWithClient creates a new bot with a provided SlackAPI implementation
func NewBotWithClient(client SlackAPI, config *Config, log *slog.Logger) (*Bot, error) {
	authTest, err := client.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Slack: %w", err)
	}

	return &Bot{
		client:          client,
		botUserID:       authTest.UserID,
		config:          config,
		activeThreads:   make(map[string]bool),
		processedEvents: make(map[string]bool),
		log:             log,
	}, nil
}

// HandleEvent processes Slack events
func (b *Bot) HandleEvent(event slackevents.EventsAPIEvent) error {
	b.log.Info(fmt.Sprintf("Received event: %s", event.Type))
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			if b.isEventProcessed(ev.Channel, ev.TimeStamp) {
				return nil
			}
			b.markEventProcessed(ev.Channel, ev.TimeStamp)

			b.log.Info(fmt.Sprintf("Received mention: %s", ev.Text))
			return b.handleMention(ev)

		case *slackevents.MessageEvent:
			if ev.ThreadTimeStamp != "" && ev.User != b.botUserID {
				if b.isEventProcessed(ev.Channel, ev.TimeStamp) {
					return nil
				}

				threadKey := fmt.Sprintf("%s-%s", ev.Channel, ev.ThreadTimeStamp)
				isBotMentioned := strings.Contains(ev.Text, fmt.Sprintf("<@%s>", b.botUserID))

				if b.activeThreads[threadKey] || isBotMentioned {
					b.markEventProcessed(ev.Channel, ev.TimeStamp)

					b.log.Info(fmt.Sprintf("Received thread reply: %s", ev.Text))
					return b.handleThreadReply(ev)
				}
			} else if ev.User != b.botUserID && strings.Contains(ev.Text, fmt.Sprintf("<@%s>", b.botUserID)) {
				if b.isEventProcessed(ev.Channel, ev.TimeStamp) {
					return nil
				}

				b.markEventProcessed(ev.Channel, ev.TimeStamp)

				b.log.Info(fmt.Sprintf("Received mention in thread: %s", ev.Text))
				return b.handleMention(&slackevents.AppMentionEvent{
					User:      ev.User,
					Text:      ev.Text,
					Channel:   ev.Channel,
					TimeStamp: ev.TimeStamp,
				})
			}
		}
	}

	return nil
}

func (b *Bot) isEventProcessed(channel, timestamp string) bool {
	b.processedEventsMutex.Lock()
	defer b.processedEventsMutex.Unlock()
	key := fmt.Sprintf("%s-%s", channel, timestamp)
	return b.processedEvents[key]
}

func (b *Bot) markEventProcessed(channel, timestamp string) {
	b.processedEventsMutex.Lock()
	defer b.processedEventsMutex.Unlock()
	key := fmt.Sprintf("%s-%s", channel, timestamp)
	b.processedEvents[key] = true
}

// handleMention processes messages that mention the bot
func (b *Bot) handleMention(ev *slackevents.AppMentionEvent) error {
	message := strings.ReplaceAll(ev.Text, fmt.Sprintf("<@%s>", b.botUserID), "")
	message = strings.TrimSpace(message)

	response := b.processMessage(message)

	_, _, err := b.client.PostMessage(
		ev.Channel,
		slack.MsgOptionText(response, false),
		slack.MsgOptionTS(ev.TimeStamp),
	)

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	threadKey := fmt.Sprintf("%s-%s", ev.Channel, ev.TimeStamp)
	b.activeThreads[threadKey] = true

	return nil
}

// handleThreadReply processes replies in threads where the bot is active
func (b *Bot) handleThreadReply(ev *slackevents.MessageEvent) error {
	response := b.processMessage(ev.Text)

	_, _, err := b.client.PostMessage(
		ev.Channel,
		slack.MsgOptionText(response, false),
		slack.MsgOptionTS(ev.ThreadTimeStamp),
	)

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	threadKey := fmt.Sprintf("%s-%s", ev.Channel, ev.ThreadTimeStamp)
	b.activeThreads[threadKey] = true

	return nil
}

// processMessage handles the actual business logic of responding to messages
func (b *Bot) processMessage(message string) string {
	return fmt.Sprintf("%s", message)
}

// CleanupOldEvents should be called periodically to prevent memory leaks
func (b *Bot) CleanupOldEvents() {
	b.processedEventsMutex.Lock()
	defer b.processedEventsMutex.Unlock()

	if len(b.processedEvents) > 1000 {
		b.processedEvents = make(map[string]bool)
	}
}
