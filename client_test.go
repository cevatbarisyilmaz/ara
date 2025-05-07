package ara_test

import (
	"bytes"
	"context"
	"github.com/cevatbarisyilmaz/ara"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
)

func TestNewClient(t *testing.T) {
	const testMessage = "Hi!"
	const host = "example.com"
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(testMessage))
		if err != nil {
			t.Fatal(err)
		}
	})}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	client := ara.NewClient(ara.NewCustomResolver(map[string][]string{host: {"127.0.0.1"}}))
	defer client.CloseIdleConnections()
	response, err := client.Get("http://" + host + ":" + strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buffer, []byte(testMessage)) {
		t.Fatal("failed")
	}
}
