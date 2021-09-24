package slack

import "github.com/pkg/errors"

var (
	errUnexpectedInnerEventData = errors.New("unexpected evt.data.inner_event.data")
	ErrInvalidCommand           = errors.New("invalid command")
)

type SlackError interface {
	Error() string
	Cause() error

	Channel() *string
}

type slackError struct {
	err     error
	channel *string
}

func (e *slackError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e *slackError) Cause() error {
	if e.err != nil {
		return e.err
	}
	return e
}

func (e *slackError) Channel() *string {
	return e.channel
}

type errOptions struct {
	channel *string
}

type ErrorOption interface {
	apply(*errOptions)
}

type channelOption string

func (s channelOption) apply(opts *errOptions) {
	opts.channel = toStrPtr(string(s))
}

func withChannel(s string) ErrorOption {
	return channelOption(s)
}

func toStrPtr(s string) *string {
	return &s
}

func NewSlackError(err error, opts ...ErrorOption) error {
	errOptions := errOptions{
		channel: nil,
	}

	for _, o := range opts {
		o.apply(&errOptions)
	}

	return &slackError{
		err:     err,
		channel: errOptions.channel,
	}
}
