package slack

import (
	"context"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
)

func SendMessage(ctx context.Context, api *slack.Client, channel, msg string) error {
	if _, _, _, err := api.SendMessageContext(ctx, channel, slack.MsgOptionText(msg, false)); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
