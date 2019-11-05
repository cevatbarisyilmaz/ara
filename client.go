package ara

import "net/http"

func NewClient(r Resolver) *http.Client {
	return &http.Client{
		Transport: NewTransport(r),
	}
}
