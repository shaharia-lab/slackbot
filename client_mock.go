package slackbot

import "github.com/slack-go/slack"

// MockSlackClient implements the SlackAPI interface for testing
type MockSlackClient struct {
	PostMessageFunc func(channelID string, options ...slack.MsgOption) (string, string, error)
	AuthTestFunc    func() (*slack.AuthTestResponse, error)

	// Track calls for verification in tests
	PostMessageCalls []PostMessageCall
}

type PostMessageCall struct {
	ChannelID string
	Options   []slack.MsgOption
}

func (m *MockSlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	m.PostMessageCalls = append(m.PostMessageCalls, PostMessageCall{
		ChannelID: channelID,
		Options:   options,
	})
	return m.PostMessageFunc(channelID, options...)
}

func (m *MockSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	return m.AuthTestFunc()
}

// NewMockSlackClient creates a new mock client with default implementations
func NewMockSlackClient() *MockSlackClient {
	return &MockSlackClient{
		PostMessageFunc: func(channelID string, options ...slack.MsgOption) (string, string, error) {
			return "message-id", "timestamp", nil
		},
		AuthTestFunc: func() (*slack.AuthTestResponse, error) {
			return &slack.AuthTestResponse{
				UserID: "mock-bot-id",
			}, nil
		},
	}
}
