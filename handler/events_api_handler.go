package handler

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"github.com/slack-go/slack/slackevents"
)

const (
	ctxTagsEventDataType     = "evt.data.type"
	ctxTagsKeyInnerEventType = "evt.data.inner_event.type"
	ctxTagsKeyInnerEventData = "evt.data.inner_event.data"
)

var errUnexpectedInnerEventData = errors.New("unexpected evt.data.inner_event.data")

type OnReactionAddedHandlerFunc = func(ctx context.Context, d *slackevents.ReactionAddedEvent) error

func EventsAPIHandler(
	ctx context.Context,
	eventsAPIEvent slackevents.EventsAPIEvent,
	onReactionAddedHandler OnReactionAddedHandlerFunc,
) error {
	tags := grpc_ctxtags.Extract(ctx)
	tags.Set(ctxTagsEventDataType, eventsAPIEvent.Type)
	tags.Set(ctxTagsKeyInnerEventType, eventsAPIEvent.InnerEvent.Type)

	switch eventsAPIEvent.InnerEvent.Type {
	case slackevents.AppMention:
	case slackevents.ReactionAdded:
		d, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.ReactionAddedEvent)
		if !ok {
			tags.Set(ctxTagsKeyInnerEventData, d)
			return errors.WithStack(errUnexpectedInnerEventData)
		}
		if onReactionAddedHandler != nil {
			return onReactionAddedHandler(ctx, d)
		}
		return nil
	}

	return errors.New("unsupported Events API event received")
}
