package slack

import (
	"context"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Client = slack.Client
type SocketClient = socketmode.Client
type AppMentionEvent = slackevents.AppMentionEvent

type SocketServer interface {
	Listen()
	Run() error
	SlackAPI() *Client

	OnReactionAdded(f onReactionAddedHandlerFunc)
	OnAppMentionCommand(command string, f onAppMentionCommandHandlerFunc)
}

type DefaultSocketServer struct {
	options options

	api    *Client
	client *SocketClient

	onReactionAddedHandler  onReactionAddedHandlerFunc
	onAppMentionCommandFunc sync.Map
}

func (s *DefaultSocketServer) Listen() {
	for evt := range s.SocketClient().Events {
		ctx := ctxlogrus.ToContext(
			grpc_ctxtags.SetInContext(context.Background(), grpc_ctxtags.NewTags()),
			logrus.WithField("evt.type", evt.Type),
		)
		err := s.handleSocketEvent(ctx, evt)
		entry := ctxlogrus.Extract(ctx).WithContext(ctx)
		if e, ok := err.(SlackError); ok && errors.Cause(err) == ErrInvalidCommand && e.Channel() != nil {
			if err := s.SendHelpMessage(ctx, *e.Channel()); err != nil {
				logrus.WithError(err).Error("send help message")
			}
		}
		if err != nil {
			entry.WithField("error", err).Error(err.Error())
			continue
		}
		entry.Info("succeeded")
	}
}

func (s *DefaultSocketServer) Run() error {
	return s.SocketClient().Run()
}

func (s *DefaultSocketServer) OnReactionAdded(f onReactionAddedHandlerFunc) {
	s.onReactionAddedHandler = f
}

func (s *DefaultSocketServer) OnAppMentionCommand(command string, f onAppMentionCommandHandlerFunc) {
	s.onAppMentionCommandFunc.Store(command, f)
}

func (s *DefaultSocketServer) onAppMentionCommandHandler(ctx context.Context, d *AppMentionEvent, command string, args []string) error {
	i, ok := s.onAppMentionCommandFunc.Load(command)
	if !ok {
		return s.SendHelpMessage(ctx, d.Channel)
	}

	f, ok := i.(onAppMentionCommandHandlerFunc)
	if !ok {
		return errors.New("unexpected func founded")
	}
	return f(ctx, d, s.SlackAPI(), args)
}

func (s *DefaultSocketServer) SocketClient() *SocketClient {
	return s.client
}

func (s *DefaultSocketServer) SlackAPI() *Client {
	return s.api
}

func (s *DefaultSocketServer) SendHelpMessage(ctx context.Context, channel string) error {
	if err := SendMessage(ctx, s.SlackAPI(), channel, s.options.helpMessage); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

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
		if err := handleEventsAPI(ctx, eventsAPIEvent, s.onReactionAddedHandler, s.onAppMentionCommandHandler); err != nil {
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

func NewSocketServer(botToken, rtmToken string, opts ...Option) SocketServer {
	options := options{
		debug:       false,
		helpMessage: "override help message required",
	}

	for _, o := range opts {
		o.apply(&options)
	}

	api := slack.New(
		botToken,
		slack.OptionDebug(options.debug),
		slack.OptionAppLevelToken(rtmToken),
	)
	client := socketmode.New(
		api,
		socketmode.OptionDebug(options.debug),
	)

	return &DefaultSocketServer{
		options:                 options,
		api:                     api,
		client:                  client,
		onReactionAddedHandler:  nil,
		onAppMentionCommandFunc: sync.Map{},
	}
}
