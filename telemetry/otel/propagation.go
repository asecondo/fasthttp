package otel

import "github.com/valyala/fasthttp"

type headerCarrier struct {
	header *fasthttp.RequestHeader
}

func (hc *headerCarrier) Get(key string) string {
	return b2s(hc.header.Peek(key))
}

func (hc *headerCarrier) Set(key string, value string) {
	hc.header.Set(key, value)
}

func (hc *headerCarrier) Keys() []string {
	// TODO (NOW): do we need to alloc a new list? this mimics what the otel library does.
	keys := make([]string, 0, hc.header.Len())
	for _, key := range hc.header.PeekKeys() {
		keys = append(keys, b2s(key))
	}
	return keys
}
