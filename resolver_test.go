package ara_test

import (
	"context"
	"github.com/cevatbarisyilmaz/ara"
	"testing"
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
