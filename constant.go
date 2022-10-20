package tracing

import "errors"

var (
	SpanPropagateCtxKey   = "opentracing-span"
	ErrInvalidSpanCarrier = errors.New("tracing: invalid span carrier")
)
