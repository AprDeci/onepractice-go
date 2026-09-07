package message_queue

type Option func(*Options)

type Options struct {
	topic   string
	handler Handler
}

func WithTopic(topic string) Option {
	return func(opts *Options) {
		opts.topic = topic
	}
}

func WithHandler(handler Handler) Option {
	return func(opts *Options) {
		opts.handler = handler
	}
}
