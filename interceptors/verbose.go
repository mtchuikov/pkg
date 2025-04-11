package interceptors

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/rs/zerolog"
)

func Verbose(logger zerolog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			resp, err := next(ctx, req)
			duration := time.Since(start)

			event := logger.Info()
			if err != nil {
				event.Err(err)
			}

			event.Str("from", req.Peer().Addr).
				Str("method", req.HTTPMethod()).
				Str("procedure", req.Spec().Procedure).
				Str("stream", req.Spec().StreamType.String()).
				Dur("duration", duration).
				Msg("request handled")

			return resp, err
		}
	}
}
