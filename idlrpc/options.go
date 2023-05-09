package idlrpc

import (
	"context"

	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/log"
)

type (
	Options struct {
		ctx        context.Context
		logger     log.ILogger
		stackTrace bool
		callTrace  bool
	}
	Option func(*Options)
)

func (o *Options) StackTrace() bool {
	return o.stackTrace
}

func (o *Options) CallTrace() bool {
	return o.callTrace
}

func WithUserData(key, val interface{}) Option {
	return func(o *Options) {
		o.ctx = context.WithValue(o.ctx, key, val)
	}
}

func WithLogger(logger log.ILogger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

func WithStackTrace(open bool) Option {
	return func(o *Options) {
		o.stackTrace = open
	}
}

func WithCallTrace(open bool) Option {
	return func(o *Options) {
		o.callTrace = open
	}
}
