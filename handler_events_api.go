package slack

import (
	"context"
	"fmt"
	"strings"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	ctxTagsEventDataType     = "evt.data.type"
	ctxTagsKeyInnerEventType = "evt.data.inner_event.type"
	ctxTagsKeyInnerEventData = "evt.data.inner_event.data"
)

var (
	errUnexpectedInnerEventData = errors.New("unexpected evt.data.inner_event.data")
	ErrInvalidCommand           = errors.New("invalid command")

	CtxChannelMarkerKey = &struct{}{}
)

type OnReactionAddedHandlerFunc = func(ctx context.Context, d *slackevents.ReactionAddedEvent) error
type OnAppMentionCommandHandlerFunc = func(ctx context.Context, d *slackevents.AppMentionEvent, api *slack.Client, args []string) error
type OnAppMentionCommandHandlerExecutor = func(ctx context.Context, d *slackevents.AppMentionEvent, command string, args []string) error

func EventsAPIHandler(
	ctx context.Context,
	eventsAPIEvent slackevents.EventsAPIEvent,
	onReactionAddedHandler OnReactionAddedHandlerFunc,
	onAppMentionCommand OnAppMentionCommandHandlerExecutor,
) error {
	tags := grpc_ctxtags.Extract(ctx)
	tags.Set(ctxTagsEventDataType, eventsAPIEvent.Type)
	tags.Set(ctxTagsKeyInnerEventType, eventsAPIEvent.InnerEvent.Type)

	switch eventsAPIEvent.InnerEvent.Type {
	case slackevents.AppMention:
		d, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			tags.Set(ctxTagsKeyInnerEventData, d)
			return errors.WithStack(errUnexpectedInnerEventData)
		}
		tags.Set(ctxTagsKeyInnerEventData, logrus.Fields{
			"user":        d.User,
			"channel":     d.Channel,
			"text":        d.Text,
			"description": fmt.Sprintf("%s User mention in channel %s with text %s", d.User, d.Channel, d.Text),
		})

		ss := strings.Split(strings.TrimSpace(d.Text), " ")
		if len(ss) < 2 {
			return errors.WithStack(ErrInvalidCommand)
		}
		return onAppMentionCommand(context.WithValue(ctx, CtxChannelMarkerKey, d.Channel), d, strings.ToLower(ss[1]), ss[2:])
	case slackevents.ReactionAdded:
		d, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.ReactionAddedEvent)
		if !ok {
			tags.Set(ctxTagsKeyInnerEventData, d)
			return errors.WithStack(errUnexpectedInnerEventData)
		}

		tags.Set(ctxTagsKeyInnerEventData, logrus.Fields{
			"user":        d.User,
			"reaction":    d.Reaction,
			"item_user":   d.ItemUser,
			"description": fmt.Sprintf("%s User react using %s on %s's message", d.User, d.Reaction, d.ItemUser),
		})

		if onReactionAddedHandler != nil {
			return onReactionAddedHandler(ctx, d)
		}
		return nil
	}

	return errors.New("unsupported Events API event received")
}