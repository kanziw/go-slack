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

	"github.com/kanziw/go-slack/handler"
)

type SocketServer interface {
	Listen()
	Run() error

	OnReactionAdded(f handler.OnReactionAddedHandlerFunc)
	OnAppMentionCommand(command string, f handler.OnAppMentionCommandHandlerFunc)
}

type DefaultSocketServer struct {
	options options

	api    *slack.Client
	client *socketmode.Client

	onReactionAddedHandler  handler.OnReactionAddedHandlerFunc
	onAppMentionCommandFunc sync.Map
}

func (s *DefaultSocketServer) Listen() {
	for evt := range s.SocketClient().Events {
		ctx := ctxlogrus.ToContext(
			grpc_ctxtags.SetInContext(context.Background(), grpc_ctxtags.NewTags()),
			logrus.WithField("evt.type", evt.Type),
		)
		err := func() error {
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
				if err := handler.EventsAPIHandler(ctx, eventsAPIEvent, s.onReactionAddedHandler, s.onAppMentionCommandHandler); err != nil {
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
		}()
		entry := ctxlogrus.Extract(ctx).WithContext(ctx)
		if err != nil {
			if errors.Cause(err) == handler.ErrInvalidCommand {
				if err := s.SendHelpMessage(ctx); err != nil {
					logrus.WithError(err).Error("send help message")
				}
			}
			entry.WithField("error", err).Error(err.Error())
			continue
		}
		entry.Info("succeeded")
	}
}

func (s *DefaultSocketServer) Run() error {
	return s.SocketClient().Run()
}

func (s *DefaultSocketServer) OnReactionAdded(f handler.OnReactionAddedHandlerFunc) {
	s.onReactionAddedHandler = f
}

func (s *DefaultSocketServer) OnAppMentionCommand(command string, f handler.OnAppMentionCommandHandlerFunc) {
	s.onAppMentionCommandFunc.Store(command, f)
}

func (s *DefaultSocketServer) onAppMentionCommandHandler(ctx context.Context, d *slackevents.AppMentionEvent, command string, args []string) error {
	i, ok := s.onAppMentionCommandFunc.Load(command)
	if !ok {
		return s.SendHelpMessage(ctx)
	}

	f, ok := i.(handler.OnAppMentionCommandHandlerFunc)
	if !ok {
		return errors.New("unexpected func founded")
	}
	return f(ctx, d, args)
}

func (s *DefaultSocketServer) SocketClient() *socketmode.Client {
	return s.client
}

func (s *DefaultSocketServer) SlackAPI() *slack.Client {
	return s.api
}

func (s *DefaultSocketServer) SendHelpMessage(ctx context.Context) error {
	channel, ok := ctx.Value(handler.CtxChannelMarkerKey).(string)
	if !ok {
		return errors.New("channel not found")
	}

	if _, _, _, err := s.api.SendMessageContext(ctx, channel, slack.MsgOptionText(s.options.helpMessage, false)); err != nil {
		return errors.WithStack(err)
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
