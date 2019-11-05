package ara_test

import (
	"context"
	"github.com/cevatbarisyilmaz/ara"
	"net"
	"testing"
)

func TestDialer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}
		err = conn.Close()
		if err != nil {
			t.Error(err)
		}
		err = listener.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	resolver := ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}})
	dialer := ara.Dialer{
		Resolver: resolver,
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := dialer.DialContext(context.Background(), "tcp", "example.com:"+port)
	if err != nil {
		t.Fatal(err)
	}
	if conn.RemoteAddr().String() != listener.Addr().String() {
		t.Fatal("connection failed")
	}
	conn, err = dialer.DialContext(context.Background(), "tcp", "google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	err = conn.Close()
	if err != nil {
		t.Error(err)
	}
}
