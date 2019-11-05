package ara_test

import (
	"context"
	"fmt"
	"github.com/cevatbarisyilmaz/ara"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

func ExampleNewClient() {
	server := &http.Server{Addr: "127.0.0.1:80", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Alo?"))
		if err != nil {
			log.Fatal(err)
		}
	})}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()
	client := ara.NewClient(ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}}))
	res, _ := client.Get("http://example.com")
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: Alo?
}

func ExampleNewTransport() {
	server := &http.Server{Addr: "127.0.0.1:80", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Alo?"))
		if err != nil {
			log.Fatal(err)
		}
	})}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()
	client := &http.Client{
		Transport: ara.NewTransport(ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}})),
	}
	res, _ := client.Get("http://example.com")
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: Alo?
}

func ExampleDialer() {
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:1919")
		if err != nil {
			log.Fatal(err)
		}
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		_, err = conn.Write([]byte("Alo?"))
		if err != nil {
			log.Fatal(err)
		}
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	dialer := ara.Dialer{
		Resolver: ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}}),
	}
	conn, err := dialer.DialContext(context.Background(), "tcp", "example.com:1919")
	if err != nil {
		log.Fatal(err)
	}
	res, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
	// Output: Alo?
}
