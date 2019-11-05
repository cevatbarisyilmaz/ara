package ara_test

import (
	"bytes"
	"context"
	"github.com/cevatbarisyilmaz/ara"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

func TestNewClient(t *testing.T) {
	const testMessage = "Hi!"
	const host = "example.com"
	listener, err := net.Listen("tcp", "127.0.0.1:80")
	if err != nil {
		t.Fatal(err)
	}
	sm := http.NewServeMux()
	sm.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(testMessage))
		if err != nil {
			t.Fatal(err)
		}
	})
	server := &http.Server{Handler: sm}
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
	response, err := client.Get("http://" + host)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buffer, []byte(testMessage)) {
		t.Fatal("failed")
	}
}
