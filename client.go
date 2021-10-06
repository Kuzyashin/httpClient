package httpClient

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"go.uber.org/zap"
)

type client struct {
	log          *zap.Logger
	cfg          *ClientConfig
	proxyManager ProxyManager
	headerFuncs  []func(*fasthttp.Request)
}

type Client interface {
	AddHeaderFun(fun func(*fasthttp.Request))
	MakeGetRequest(ctx context.Context, url string, ignoreBody bool, opts ...FastHttpCliOpt) (cb []byte, status int, err error)
	MakeHeadRequest(ctx context.Context, url string) (status int, err error)
}

func NewClient(log *zap.Logger, config *ClientConfig, opts ...CliOpt) Client {
	logger := log.Named("http_client")
	cli := &client{log: logger, cfg: config}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}

func (c *client) AddHeaderFun(fun func(*fasthttp.Request)) {
	c.headerFuncs = append(c.headerFuncs, fun)
}

func (c *client) getClient() *fasthttp.Client {
	cli := fasthttp.Client{
		ReadTimeout:                   c.cfg.FastHTTPConfig.ReadTimeout,
		WriteTimeout:                  c.cfg.FastHTTPConfig.WriteTimeout,
		NoDefaultUserAgentHeader:      c.cfg.FastHTTPConfig.NoDefaultUserAgentHeader,
		DisableHeaderNamesNormalizing: c.cfg.FastHTTPConfig.DisableHeaderNamesNormalizing,
		DisablePathNormalizing:        c.cfg.FastHTTPConfig.DisablePathNormalizing,

		MaxConnsPerHost:     c.cfg.FastHTTPConfig.MaxConnsPerHost,
		MaxIdleConnDuration: c.cfg.FastHTTPConfig.MaxIdleConnDuration,
		MaxConnDuration:     c.cfg.FastHTTPConfig.MaxConnDuration,
		DialDualStack:       c.cfg.FastHTTPConfig.DialDualStack,
		TLSConfig:           c.cfg.FastHTTPConfig.TLSConfig,
	}
	if c.proxyManager != nil {
		proxy := c.proxyManager.GetProxy()
		cli.Dial = c.getDialer(proxy)
	}
	return &cli
}

func (c *client) getDialer(proxy string) fasthttp.DialFunc {
	if strings.Contains(proxy, "socks") {
		return FasthttpSocksDialerWithTimeout(proxy, c.cfg.ProxyDialTimeOut)
	} else {
		return fasthttpproxy.FasthttpHTTPDialerTimeout(proxy, c.cfg.ProxyDialTimeOut)
	}
}

func (c *client) MakeGetRequest(ctx context.Context, url string, ignoreBody bool, opts ...FastHttpCliOpt) (cb []byte, status int, err error) {
	return c.makeRequest(ctx, url, ignoreBody, 0, opts...)
}

func (c *client) MakeHeadRequest(ctx context.Context, url string) (status int, err error) {
	return c.makeHeadRequest(ctx, url, 0)
}

func (c *client) makeRequest(ctx context.Context, url string, ignoreBody bool, retries int, opts ...FastHttpCliOpt) (cb []byte, status int, err error) {
	if retries > c.cfg.MaxRetries {
		return nil, 0, fmt.Errorf("max retries count reached")
	}
	cli := c.getClient()
	for _, opt := range opts {
		opt(cli)
	}
	defer cli.CloseIdleConnections()
	var body []byte
	req := fasthttp.AcquireRequest()
	req.SetConnectionClose()
	for _, fun := range c.headerFuncs {
		fun(req)
	}
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	if ignoreBody {
		resp.SkipBody = true
	}
	err = cli.Do(req, resp)
	fasthttp.ReleaseRequest(req)
	select {
	case <-ctx.Done():
		return cb, 0, nil
	default:
	}

	if err != nil {
		return c.makeRequest(ctx, url, ignoreBody, retries+1)
	}
	if !ignoreBody {
		contentEncoding := resp.Header.Peek("Content-Encoding")
		if bytes.EqualFold(contentEncoding, []byte("gzip")) {
			body, _ = resp.BodyGunzip()
		} else {
			body = resp.Body()
		}
		cb = make([]byte, len(body))
		copy(cb, body)
	}
	status = resp.StatusCode()
	for _, code := range c.cfg.RetryOnStatus {
		if code == status {
			return c.makeRequest(ctx, url, ignoreBody, retries+1)
		}
	}
	fasthttp.ReleaseResponse(resp)
	return cb, status, nil

}

func (c *client) makeHeadRequest(ctx context.Context, url string, retries int) (code int, err error) {
	if retries > c.cfg.MaxRetries {
		return 0, fmt.Errorf("max retries count reached")
	}
	cli := c.getClient()
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	for _, fun := range c.headerFuncs {
		fun(req)
	}
	req.Header.SetMethod("HEAD")
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	resp.SkipBody = true
	err = cli.Do(req, resp)
	select {
	case <-ctx.Done():
		return 0, nil
	default:
		if err != nil {
			return c.makeHeadRequest(ctx, url, retries+1)
		}
	}
	status := resp.StatusCode()
	fasthttp.ReleaseResponse(resp)
	return status, nil
}
