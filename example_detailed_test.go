package ara_test

import (
	"context"
	"fmt"
	"github.com/cevatbarisyilmaz/ara"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

type resolver struct{}

func (r resolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	// Always return the localhost
	return []string{"127.0.0.1"}, nil
}

func ExampleDetailed() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	sm := http.NewServeMux()
	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(r.URL.Path))
		if err != nil {
			log.Fatal(err)
		}
	})
	go func() {
		log.Fatal(http.Serve(listener, sm))
	}()
	r := &resolver{}
	dialer := &ara.Dialer{
		Timeout:  time.Minute,
		Resolver: r,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := http.Client{
		Transport: transport,
		Timeout:   time.Minute,
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	res, err := client.Get("http://example.com:" + port + "/mysecretpath")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
	// Output: /mysecretpath
}
