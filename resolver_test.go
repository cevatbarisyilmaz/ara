package ara_test

import (
	"context"
	"net/http/httptrace"
	"testing"

	"github.com/cevatbarisyilmaz/ara"
)

func TestNewCustomResolver(t *testing.T) {
	resolver := ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}})
	addrs, err := resolver.LookupHost(context.Background(), "example.com")
	if err != nil {
		t.Error(err)
	} else if addrs == nil {
		t.Error("addresses are nil")
	} else if len(addrs) == 0 {
		t.Error("no addresses")
	} else if len(addrs) > 1 {
		t.Error("too many addresses")
	} else if addrs[0] != "127.0.0.1" {
		t.Error("wrong address")
	}
	addrs, err = resolver.LookupHost(context.Background(), "google.com")
	if err != nil {
		t.Error(err)
	} else if addrs == nil {
		t.Error("addresses are nil")
	} else if len(addrs) == 0 {
		t.Error("no addresses")
	}
}

func TestClientTrace(t *testing.T) {
	resolver := ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}})

	var gotDNSStart, gotDNSDone bool
	ctx := context.Background()
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			if info.Host != "example.com" {
				t.Error("wrong address in ClientTrace.DNSStart")
			}
			gotDNSStart = true
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if info.Err != nil {
				t.Error("non-nil error in ClientTrace.DNSDone")
			}
			if len(info.Addrs) != 1 || info.Addrs[0].String() != "127.0.0.1" {
				t.Error("wrong IP in ClientTrace.DNSDone")
			}
			gotDNSDone = true
		},
	})

	resolver.LookupHost(ctx, "example.com")

	if !gotDNSStart {
		t.Error("ClientTrace.DNSStart not called")
	}
	if !gotDNSDone {
		t.Error("ClientTrace.DNSDone not called")
	}
}
