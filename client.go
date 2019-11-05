package ara

import "net/http"

// NewClient returns a *http.Client that uses the given resolver while dialing.
func NewClient(r Resolver) *http.Client {
	return &http.Client{
		Transport: NewTransport(r),
	}
}
