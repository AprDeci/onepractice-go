package message_queue

type Option func(*Options)

type Options struct {
	topic     string
	handler   Handler
	workers   int
	batchSize int
}

func WithWorkers(workers int) Option {
	return func(opts *Options) {
		if workers > 0 {
			opts.workers = workers
		}
	}
}
func WithBatchSize(batchSize int) Option {
	return func(opts *Options) {
		if batchSize > 0 {
			opts.batchSize = batchSize
		}
	}
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
