package slackbot_test

import (
	"github.com/shaharia-lab/slackbot"
	"log/slog"
	"os"
	"testing"

	"github.com/slack-go/slack/slackevents"
)

func TestBotHandleMention(t *testing.T) {
	mockClient := slackbot.NewMockSlackClient()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	bot, err := slackbot.NewBotWithClient(mockClient, &slackbot.Config{
		SlackBotUserOAuthToken: "mock-token",
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	event := slackevents.EventsAPIEvent{
		Type: slackevents.CallbackEvent,
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Data: &slackevents.AppMentionEvent{
				User:      "test-user",
				Text:      "<@mock-bot-id> hello bot",
				Channel:   "test-channel",
				TimeStamp: "1234567890.123456",
			},
		},
	}

	err = bot.HandleEvent(event)
	if err != nil {
		t.Fatalf("Failed to handle event: %v", err)
	}

	if len(mockClient.PostMessageCalls) != 1 {
		t.Fatalf("Expected 1 PostMessage call, got %d", len(mockClient.PostMessageCalls))
	}

	call := mockClient.PostMessageCalls[0]
	if call.ChannelID != "test-channel" {
		t.Errorf("Expected channel ID to be 'test-channel', got '%s'", call.ChannelID)
	}
}
