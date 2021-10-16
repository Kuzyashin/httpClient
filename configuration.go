package httpClient

import (
	"crypto/tls"
	"time"
)

type ClientConfig struct {
	// MaxRetries maximum count of request retries on errors. Default 0
	MaxRetries int

	// ProxyDialTimeOut timeout to dial proxy
	ProxyDialTimeOut time.Duration

	// RetryOnStatus when set with MaxRetries, client will retry request
	// after receiving specified status code
	RetryOnStatus []int

	FastHTTPConfig FastHTTPConfig
}

type FastHTTPConfig struct {
	// NoDefaultUserAgentHeader when set to true, causes the default
	// User-Agent header to be excluded from the Request.
	NoDefaultUserAgentHeader bool

	// Attempt to connect to both ipv4 and ipv6 addresses if set to true.
	//
	// This option is used only if default TCP dialer is used,
	// i.e. if Dial is blank.
	//
	// By default, client connects only to ipv4 addresses,
	// since unfortunately ipv6 remains broken in many networks worldwide :)
	DialDualStack bool

	// TLS config for https connections.
	//
	// Default TLS config is used if not set.
	TLSConfig *tls.Config

	// Maximum number of connections per each host which may be established.
	//
	// DefaultMaxConnsPerHost is used if not set.
	MaxConnsPerHost int

	// Idle keep-alive connections are closed after this duration.
	//
	// By default, idle connections are closed
	// after DefaultMaxIdleConnDuration.
	MaxIdleConnDuration time.Duration

	// Keep-alive connections are closed after this duration.
	//
	// By default, connection duration is unlimited.
	MaxConnDuration time.Duration

	// Per-connection buffer size for responses' reading.
	// This also limits the maximum header size.
	//
	// Default buffer size is used if 0.
	ReadBufferSize int

	// Per-connection buffer size for requests' writing.
	//
	// Default buffer size is used if 0.
	WriteBufferSize int

	// Maximum duration for full response reading (including body).
	//
	// By default, response read timeout is unlimited.
	ReadTimeout time.Duration

	// Maximum duration for full request writing (including body).
	//
	// By default request write timeout is unlimited.
	WriteTimeout time.Duration

	// Maximum response body size.
	//
	// The client returns ErrBodyTooLarge if this limit is greater than 0
	// and response body is greater than the limit.
	//
	// By default, response body size is unlimited.
	MaxResponseBodySize int

	// Header names are passed as-is without normalization
	// if this option is set.
	//
	// Disabled header names' normalization may be useful only for proxying
	// responses to other clients expecting case-sensitive
	// header names. See https://github.com/valyala/fasthttp/issues/57
	// for details.
	//
	// By default, request and response header names are normalized, i.e.
	// The first letter and the first letters following dashes
	// are uppercased, while all the other letters are lowercased.
	// Examples:
	//
	//     * HOST -> Host
	//     * content-type -> Content-Type
	//     * cONTENT-lenGTH -> Content-Length
	DisableHeaderNamesNormalizing bool

	// Path values are sent as-is without normalization
	//
	// Disabled path normalization may be useful for proxying incoming requests
	// to servers that are expecting paths to be forwarded as-is.
	//
	// By default, path values are normalized, i.e.
	// extra slashes are removed, special characters are encoded.
	DisablePathNormalizing bool

	// Maximum duration for waiting for a free connection.
	//
	// By default, will not waiting, return ErrNoFreeConns immediately
	MaxConnWaitTimeout time.Duration
}
