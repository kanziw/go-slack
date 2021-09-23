package slack

type options struct {
	debug bool
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
