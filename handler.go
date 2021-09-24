package slack

import (
	"context"

	"github.com/pkg/errors"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (s *DefaultSocketServer) handleSocketEvent(ctx context.Context, evt socketmode.Event) error {
	switch evt.Type {
	case socketmode.RequestTypeHello:
	case socketmode.EventTypeConnecting:
		s.SocketClient().Debugln("Connecting to Slack with Socket Mode...")
	case socketmode.EventTypeConnectionError:
		s.SocketClient().Debugln("debug", "Connection failed. Retrying later...")
	case socketmode.EventTypeConnected:
		s.SocketClient().Debugln("debug", "Connected to Slack with Socket Mode.")
	case socketmode.EventTypeEventsAPI:
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return errors.New("unknown event type:" + string(evt.Type))
		}
		s.SocketClient().Ack(*evt.Request)
		if err := EventsAPIHandler(ctx, eventsAPIEvent, s.onReactionAddedHandler, s.onAppMentionCommandHandler); err != nil {
			s.SocketClient().Debugf(err.Error())
			return err
		}
	// TODO
	case socketmode.EventTypeInteractive:
	case socketmode.EventTypeSlashCommand:
	default:
		return errors.New("unexpected event type received")
	}
	return nil
}
