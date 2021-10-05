package httpClient

import (
	"time"

	"github.com/valyala/fasthttp"
)

type FastHttpCliOpt func(client *fasthttp.Client)

func WithReadTimeout(duration time.Duration) FastHttpCliOpt {
	return func(client *fasthttp.Client) {
		client.ReadTimeout = duration
	}
}
