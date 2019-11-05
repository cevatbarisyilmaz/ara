package ara

import (
	"context"
	"net"
)

// A Resolver looks up hosts.
type Resolver interface {
	LookupHost(ctx context.Context, host string) ([]string, error)
}

type resolver struct {
	hosts map[string][]string
}

// NewCustomResolver returns a resolver that will give priority
// to given host/ip mappings on lookups.
//
// If a host is not part of the given mapping, it will use the
// net.DefaultResolver
//
// hosts is a map of addresses for a host name, like map[host][]address
func NewCustomResolver(hosts map[string][]string) Resolver {
	return &resolver{
		hosts: hosts,
	}
}

func (r *resolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	records := r.hosts[host]
	if records != nil && len(records) != 0 {
		return records, nil
	}
	return net.DefaultResolver.LookupHost(ctx, host)
}
