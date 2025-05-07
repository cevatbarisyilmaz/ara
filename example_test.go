package ara_test

import (
	"context"
	"fmt"
	"github.com/cevatbarisyilmaz/ara"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

func ExampleNewClient() {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	err = listener.Close() // free the port for the server
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{
		Addr: listener.Addr().String(),
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, err := writer.Write([]byte("Alo?"))
			if err != nil {
				log.Fatal(err)
			}
		})}
	go func() {
		err = server.ListenAndServe()
		log.Println(err)
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()
	client := ara.NewClient(ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}}))
	defer client.CloseIdleConnections()

	res, err := client.Get("http://example.com:" + strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: Alo?
}

func ExampleNewTransport() {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	err = listener.Close() // free the port for the server
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	var server *http.Server
	go func() {
		addr := listener.Addr().String()
		server = &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("Alo?"))
			if err != nil {
				log.Fatal(err)
			}
		})}
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
	defer client.CloseIdleConnections()
	res, err := client.Get("http://example.com:" + strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	// Output: Alo?
}

func ExampleDialer() {
	serverUp := make(chan bool, 1)
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:1919")
		if err != nil {
			log.Fatal(err)
		}
		serverUp <- true
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
	<-serverUp
	dialer := ara.Dialer{
		Resolver: ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}}),
	}
	conn, err := dialer.DialContext(context.Background(), "tcp", "example.com:1919")
	if err != nil {
		log.Fatal(err)
	}
	res, err := io.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
	// Output: Alo?
}
