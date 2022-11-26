package middleware

import (
	"context"

	"github.com/Xuwudong/myraft/logger"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/google/uuid"
)

func TrackIdProcessorMiddleware() thrift.ProcessorMiddleware {
	return func(name string, next thrift.TProcessorFunction) thrift.TProcessorFunction {
		return thrift.WrappedTProcessorFunction{
			Wrapped: func(ctx context.Context, seqId int32, in, out thrift.TProtocol) (bool, thrift.TException) {
				v := ctx.Value(logger.TraceIDKey)
				traceID, _ := v.(string)
				if traceID == "" {
					uid := uuid.New()
					traceID = uid.String()
					ctx = context.WithValue(ctx, logger.TraceIDKey, traceID)
				}
				return next.Process(ctx, seqId, in, out)
			},
		}
	}
}
