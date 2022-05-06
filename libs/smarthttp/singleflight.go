package smarthttp

import (
	"net/http"
	"strings"

	"golang.org/x/sync/singleflight"
)

// Singleflight defines the Singleflight configuration
type Singleflight struct {
	// KeyGenerator will generate a unique "key" from the request.
	// This key will be used to deduplicate requests.
	// If none is provided then DefaultSFKeyGenerator is used.
	KeyGenerator func(req *http.Request) string

	group              *singleflight.Group
	actualKeyGenerator func(req *http.Request) string

	instrumentation Instrumentation

	// used for testing only
	trackKey    func(s *Singleflight, key string)
	trackedKeys []string
}

func (s *Singleflight) buildMiddleware(doFunc requestClosure) requestClosure {
	return func(req *http.Request) (*http.Response, error) {
		// disable singleflight for non-GET requests when a custom key generator was not supplied
		if s.KeyGenerator == nil && req.Method != http.MethodGet {
			return doFunc(req)
		}

		key := s.actualKeyGenerator(req)
		s.trackKey(s, key)

		var innerErr error

		//nolint:bodyclose
		result, err, _ := s.group.Do(key, func() (interface{}, error) {
			var resp interface{}
			resp, innerErr = doFunc(req)

			return resp, innerErr
		})

		if err != nil && innerErr == nil {
			s.instrumentation.SingleflightErr(req, err)
		}

		if result != nil {
			return result.(*http.Response), err
		}

		return nil, err
	}
}

func (s *Singleflight) addMiddleware(doFunc requestClosure) requestClosure {
	if s == nil {
		return doFunc
	}

	return s.buildMiddleware(doFunc)
}

func (s *Singleflight) doInitOnce(instrumentation Instrumentation) {
	if s == nil {
		instrumentation.InitWarning("no single flight has been configured.  Use is strongly recommended for all read requests")

		return
	}

	s.instrumentation = instrumentation

	s.group = &singleflight.Group{}

	if s.KeyGenerator != nil {
		s.actualKeyGenerator = s.KeyGenerator
	} else {
		s.actualKeyGenerator = DefaultSFKeyGenerator
	}

	if s.trackKey == nil {
		s.trackKey = func(_ *Singleflight, _ string) {
			// noop
		}
	}
}

// DefaultSFKeyGenerator generate a "unique key" from the request content (see implementation for details).
// NOTE: this implementation does not look at the body and as such is only appropriate for properly formed GET requests (e.g. no body)
func DefaultSFKeyGenerator(req *http.Request) string {
	builder := strings.Builder{}

	_, _ = builder.WriteString(req.Method)
	_, _ = builder.Write([]byte(`||`))
	_, _ = builder.WriteString(req.URL.String())

	return builder.String()
}
