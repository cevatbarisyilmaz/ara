package ara

import (
	"context"
	"net"
	"net/http/httptrace"
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
// net.DefaultResolver.
//
// hosts is a map of addresses for a host name, like map[host][]address.
func NewCustomResolver(hosts map[string][]string) Resolver {
	return &resolver{
		hosts: hosts,
	}
}

func handleClientTrace(t *httptrace.ClientTrace, host string, records []string) {
	if t.DNSStart != nil {
		t.DNSStart(httptrace.DNSStartInfo{
			Host: host,
		})
	}
	if t.DNSDone != nil {
		var addrs []net.IPAddr
		for _, rec := range records {
			ip := net.ParseIP(rec)
			if ip != nil {
				addrs = append(addrs, net.IPAddr{
					IP: ip,
				})
			}
		}
		t.DNSDone(httptrace.DNSDoneInfo{
			Addrs: addrs,
		})
	}
}

func (r *resolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	records := r.hosts[host]
	if records != nil && len(records) != 0 {
		t := httptrace.ContextClientTrace(ctx)
		if t != nil {
			handleClientTrace(t, host, records)
		}

		return records, nil
	}
	return net.DefaultResolver.LookupHost(ctx, host)
}
