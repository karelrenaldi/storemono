package smarthttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// This is the default timeout set into the default HTTP client (if a client is not supplied)
	defaultTimeout = 3 * time.Second
	// This is the default connect timeout set into the
	defaultConnectTimeout = 1 * time.Second
)

var (
	// ErrConnectTimeout indicates that we were unable to connect to the destination host and by extension the destination host cannot have
	// processed this request in any way
	ErrConnectTimeout = errors.New("connection timeout")

	// ErrConnection indicates that there were errors (other than timeout) connecting to the destination host
	ErrConnection = errors.New("error initiating connection")

	// ErrTimeout indicates that we succeeded to connect to the destination host but failed to receive the response before the Timeout or
	// context timeout expired.
	// By extension this error implies that the destination received the request and may have partial processed it.
	ErrTimeout = errors.New("timeout")
)

// Client is a drop-in replacement for the standard http.Client that provides additional features.
// Please note, as with the http.Client it is strongly recommended that a single instance of this client is created and then
// shared amongst the goroutines that make this request type.  Allowing for connection pooling and other performance optimizations.
type Client struct {
	// Name is the unique name for this client.
	// This name is used to track errors, emit stats, etc.
	// It is recommended to use an identifiable name link the service or endpoint being called.
	Name string

	// Client is the underlying HTTP client that will be used to make the requests.
	// User are encouraged to populate this and explicitly set timeouts.
	// If users do not populate this field, it will be automatically populated with this package's default settings.
	// Users must not change or access this client after initial creation and a data race may result.
	Client         *http.Client
	clientInitOnce sync.Once

	// Timeout is the total timeout (including connection and read timeout) of a particular request
	Timeout time.Duration

	// ConnectTimeout is the timeout for the connection initiation phase.
	// Note: ConnectTimeout should be lesser than timeout. Else, ErrConnectTimeout cannot be caught
	ConnectTimeout time.Duration

	// Instrumentation allows reporting and logging of internal events and statistics
	Instrumentation Instrumentation

	// CircuitBreaker defines the (optional) circuit breaker configuration for this client.
	CircuitBreaker CircuitBreaker

	// Retries defines the (optional) retry configuration for this client.
	Retries *Retries

	// Singleflight defines the (optional) single-flight configuration for this client.
	Singleflight *Singleflight
}

// Do performs the HTTP request provided.
//
// Note: This method does not take a context as it uses the context inside the Request parameter.
// Note: Timeouts should be set using the context.Context in the Request.
// For more information see https://godoc.org/net/http#Client.Do
// nolint:funlen
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()
	path := c.getInstrumentation().SanitizePath(req.URL.Path)
	endpointTag := generateEndpointTag(req.Method, path)

	defer c.getInstrumentation().DoDuration(start, endpointTag)

	// base request
	doRequestFunc := func(req *http.Request) (*http.Response, error) {
		resp, err := c.getClient().Do(req)
		if err != nil {
			c.getInstrumentation().BaseDoDuration(start, 0, endpointTag)

			var urlErr *url.Error

			switch {
			case errors.As(err, &urlErr) && urlErr.Timeout():
				c.getInstrumentation().BaseDoErr(err, endpointTag, "timeout")
				return resp, fmt.Errorf("%w - %s", ErrTimeout, err)

			case errors.Is(err, context.DeadlineExceeded):
				c.getInstrumentation().BaseDoErr(err, endpointTag, "ctxTimeout")
				return resp, err

			case errors.Is(err, context.Canceled):
				c.getInstrumentation().BaseDoErr(err, endpointTag, "ctxCanceled")
				return resp, err

			default:
				c.getInstrumentation().BaseDoErr(err, endpointTag, "na")
				return resp, err
			}
		}

		c.getInstrumentation().BaseDoDuration(start, resp.StatusCode, endpointTag)

		return resp, nil
	}

	// add middleware (note: be wary of the ordering here)

	// retries are inside the circuit; this means the circuit only see complete failure
	doRequestFunc = c.Retries.addMiddleware(doRequestFunc)
	doRequestFunc = (&c.CircuitBreaker).addMiddleware(doRequestFunc)

	// singleflight is last so that it does not see or interact with the retries
	doRequestFunc = c.Singleflight.addMiddleware(doRequestFunc)

	// perform request + middleware
	resp, err := doRequestFunc(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// all access to the http.Client by this struct should be via this method.
func (c *Client) getClient() *http.Client {
	c.clientInitOnce.Do(c.doInitOnce)

	return c.Client
}

// all access to the Instrumentation by this struct should be via this method.
func (c *Client) getInstrumentation() Instrumentation {
	c.clientInitOnce.Do(c.doInitOnce)

	return c.Instrumentation
}

func (c *Client) doInitOnce() {
	if c.Instrumentation == nil {
		c.Instrumentation = &noopInstrumentation{}
	}

	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}

	if c.Client == nil {
		c.Client = buildClient(c.Timeout, c.ConnectTimeout)
	}

	if c.Name == "" {
		c.Instrumentation.InitWarning("name was not supplied.  Use of unique and informative names is strongly recommended")

		c.Name = fmt.Sprintf("smart-http-%d", time.Now().UnixNano())
	}

	c.Instrumentation.Init(c.Name)

	(&c.CircuitBreaker).doInitOnce(c.Instrumentation, c.Name)

	if c.Retries != nil {
		c.Retries.doInitOnce(c.Instrumentation)
	}

	if c.Singleflight != nil {
		c.Singleflight.doInitOnce(c.Instrumentation)
	}
}

// GetTransportWithCustomDialer is used internally to assist with detecting connection timeouts during Dial().
// It is provided here so others can use it with their own http.Transport.
func GetTransportWithCustomDialer(connectionTimeout time.Duration) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			dialer := net.Dialer{
				Timeout: connectionTimeout,
			}

			conn, err = dialer.DialContext(ctx, network, addr)
			if err != nil {
				if netError, ok := err.(net.Error); ok {
					if netError.Timeout() {
						return nil, ErrConnectTimeout
					}
					return nil, fmt.Errorf("%w %v", ErrConnection, err)
				}

				return nil, err
			}

			return conn, nil
		},
	}
}

func buildClient(timeout, connectTimeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: GetTransportWithCustomDialer(connectTimeout),
	}
}

func generateEndpointTag(method, path string) string {
	return method + "::" + path
}

type requestClosure func(*http.Request) (*http.Response, error)
