package httpClient

import (
	"net"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
	"h12.io/socks"
)

// FasthttpSocksDialerWithTimeout returns a fasthttp.DialFunc that dials using
// the provided SOCKS5 proxy with timeout.
//
// Example usage:
//	c := &fasthttp.Client{
//		Dial: fasthttpproxy.FasthttpSocksDialerWithTimeout("socks5://localhost:9050", time.second * 3),
//	}

func FasthttpSocksDialerWithTimeout(proxyAddr string, timeout time.Duration) fasthttp.DialFunc {

	return func(addr string) (net.Conn, error) {
		dialer := socks.Dial(proxyAddr + "?timeout=" + timeout.String())
		return dialer("tcp", addr)
	}
}

// FasthttpSocksDialer returns a fasthttp.DialFunc that dials using
// the provided SOCKS5 proxy.
//
// Example usage:
//	c := &fasthttp.Client{
//		Dial: fasthttpproxy.FasthttpSocksDialer("socks5://localhost:9050"),
//	}
func FasthttpSocksDialer(proxyAddr string) fasthttp.DialFunc {
	var (
		u      *url.URL
		err    error
		dialer proxy.Dialer
	)
	if u, err = url.Parse(proxyAddr); err == nil {
		dialer, err = proxy.FromURL(u, proxy.Direct)
	}
	// It would be nice if we could return the error here. But we can't
	// change our API so just keep returning it in the returned Dial function.
	// Besides the implementation of proxy.SOCKS5() at the time of writing this
	// will always return nil as error.

	return func(addr string) (net.Conn, error) {
		if err != nil {
			return nil, err
		}
		return dialer.Dial("tcp", addr)
	}
}
