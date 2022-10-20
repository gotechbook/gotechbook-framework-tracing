package tracing

import (
	"context"
	"errors"
	"github.com/gotechbook/gotechbook-framework-tracing/jaeger"
	"io"
	"os"
	"testing"

	gContext "github.com/gotechbook/gotechbook-framework-context"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

var closer io.Closer

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	closer, _ = jaeger.Configure(jaeger.Options{ServiceName: "spanTest"})
}

func shutdown() {
	closer.Close()
}

func assertBaggage(t *testing.T, ctx opentracing.SpanContext, expected map[string]string) {
	b := extractBaggage(ctx, true)
	assert.Equal(t, expected, b)
}

func extractBaggage(ctx opentracing.SpanContext, allItems bool) map[string]string {
	b := make(map[string]string)
	ctx.ForeachBaggageItem(func(k, v string) bool {
		b[k] = v
		return allItems
	})
	return b
}

func TestExtractSpan(t *testing.T) {
	span := opentracing.StartSpan("op", opentracing.ChildOf(nil))
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	spanCtx, err := ExtractSpan(ctx)
	assert.NoError(t, err)
	assert.Equal(t, span.Context(), spanCtx)
}

func TestExtractSpanInjectedSpan(t *testing.T) {
	span := opentracing.StartSpan("someOp")
	span.SetBaggageItem("some_key", "12345")
	span.SetBaggageItem("some-other-key", "42")
	expectedBaggage := map[string]string{"some_key": "12345", "some-other-key": "42"}

	spanData := opentracing.TextMapCarrier{}
	tracer := opentracing.GlobalTracer()
	err := tracer.Inject(span.Context(), opentracing.TextMap, spanData)
	assert.NoError(t, err)
	ctx := gContext.AddToPropagateCtx(context.Background(), SpanPropagateCtxKey, spanData)

	spanCtx, err := ExtractSpan(ctx)
	assert.NoError(t, err)
	assertBaggage(t, spanCtx, expectedBaggage)
}

func TestExtractSpanNoSpan(t *testing.T) {
	spanCtx, err := ExtractSpan(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, spanCtx)
}

func TestExtractSpanBadInjected(t *testing.T) {
	ctx := gContext.AddToPropagateCtx(context.Background(), SpanPropagateCtxKey, []byte("nope"))
	spanCtx, err := ExtractSpan(ctx)
	assert.Equal(t, ErrInvalidSpanCarrier, err)
	assert.Nil(t, spanCtx)
}

func TestInjectSpanContextWithoutSpan(t *testing.T) {
	origCtx := context.Background()
	ctx, err := InjectSpan(origCtx)
	assert.NoError(t, err)
	assert.Equal(t, origCtx, ctx)
}

func TestInjectSpan(t *testing.T) {
	span := opentracing.StartSpan("op", opentracing.ChildOf(nil))
	origCtx := opentracing.ContextWithSpan(context.Background(), span)
	ctx, err := InjectSpan(origCtx)
	assert.NoError(t, err)
	assert.NotEqual(t, origCtx, ctx)
	encodedCtx := gContext.GetFromPropagateCtx(ctx, SpanPropagateCtxKey)
	assert.NotNil(t, encodedCtx)
}

func TestStartSpan(t *testing.T) {
	origCtx := context.Background()
	ctxWithSpan := StartSpan(origCtx, "my-op", opentracing.Tags{"hi": "hello"})
	assert.NotEqual(t, origCtx, ctxWithSpan)

	span := opentracing.SpanFromContext(ctxWithSpan)
	assert.NotNil(t, span)
}

func TestFinishSpanNilCtx(t *testing.T) {
	assert.NotPanics(t, func() { FinishSpan(nil, nil) })
}

func TestFinishSpanCtxWithoutSpan(t *testing.T) {
	assert.NotPanics(t, func() { FinishSpan(context.Background(), nil) })
}

func TestFinishSpanWithErr(t *testing.T) {
	ctxWithSpan := StartSpan(context.Background(), "my-op", opentracing.Tags{"hi": "hello"})
	assert.NotPanics(t, func() { FinishSpan(ctxWithSpan, errors.New("hello")) })
}

func TestFinishSpan(t *testing.T) {
	ctxWithSpan := StartSpan(context.Background(), "my-op", opentracing.Tags{"hi": "hello"})
	assert.NotPanics(t, func() { FinishSpan(ctxWithSpan, nil) })
}
