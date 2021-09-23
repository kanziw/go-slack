package slack

type options struct {
	debug       bool
	helpMessage string
}

type Option interface {
	apply(*options)
}

type debugOption bool

func (b debugOption) apply(opts *options) {
	opts.debug = bool(b)
}

func WithDebug(c bool) Option {
	return debugOption(c)
}

type helpMessageOption string

func (s helpMessageOption) apply(opts *options) {
	opts.helpMessage = string(s)
}

func WithHelpMessage(s string) Option {
	return helpMessageOption(s)
}
