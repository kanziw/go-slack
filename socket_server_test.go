package slack

import (
	"context"
	"testing"

	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
)

const (
	testUser    = "kanziw"
	testChannel = "channel"
)

func TestAppMentionCommand(t *testing.T) {
	cases := []struct {
		name                   string
		command                string
		expectedErr            error
		additionalErrCheckFunc func(err error)
	}{
		{
			name:        "only mention",
			command:     testUser,
			expectedErr: ErrInvalidCommand,
			additionalErrCheckFunc: func(err error) {
				e, ok := err.(SlackError)
				assert.True(t, ok)
				assert.NotNil(t, e.Channel())
				assert.Equal(t, testChannel, *e.Channel())
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			evt := appMentionEvtWithCommand(tc.command)

			err := handleEventsAPI(ctx, evt, nil, nil)
			assert.Error(t, err, tc.expectedErr.Error())

			if tc.additionalErrCheckFunc != nil {
				tc.additionalErrCheckFunc(err)
			}
		})
	}
}

func appMentionEvtWithCommand(command string) slackevents.EventsAPIEvent {
	return slackevents.EventsAPIEvent{
		APIAppID: "",
		Data:     nil,
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Type: slackevents.AppMention,
			Data: &AppMentionEvent{
				User:    testUser,
				Text:    command,
				Channel: testChannel,
			},
		},
	}
}
