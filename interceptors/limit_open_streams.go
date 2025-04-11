package interceptors

import (
	"context"
	"sync/atomic"

	"connectrpc.com/connect"
)

type limitOpenStreams struct {
	max     int64
	current int64
}

func LimitOpenStreams(max int) connect.Interceptor {
	return &limitOpenStreams{
		max:     int64(max),
		current: 0,
	}
}

func (l *limitOpenStreams) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return next(ctx, req)
	}
}

func (l *limitOpenStreams) WrapStreamingClient(
	next connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (l *limitOpenStreams) WrapStreamingHandler(
	next connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if atomic.LoadInt64(&l.current) >= l.max {
			return connect.NewError(connect.CodeResourceExhausted,
				ErrMaxOpenStreamsReached)
		}

		atomic.AddInt64(&l.current, 1)
		handler := next(ctx, conn)
		atomic.AddInt64(&l.current, -1)

		return handler
	}
}
