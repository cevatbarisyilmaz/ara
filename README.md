# ara
[![GoDoc](https://godoc.org/github.com/cevatbarisyilmaz/ara?status.svg)](https://godoc.org/github.com/cevatbarisyilmaz/ara)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cevatbarisyilmaz/ara?sort=semver)](https://github.com/cevatbarisyilmaz/ara/releases)
[![GitHub](https://img.shields.io/github/license/cevatbarisyilmaz/ara)](https://github.com/cevatbarisyilmaz/ara/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/cevatbarisyilmaz/ara)](https://goreportcard.com/report/github.com/cevatbarisyilmaz/ara)

Package ara provides a dialer with customizable resolver.

It can be used with `http.Client` and `http.Transport` to alter host lookups.
For example, with a custom resolver that maps `google.com` to `127.0.0.1`,
you can get `httpClient.Get("http://google.com")` to connect the `localhost`.

## Example

```go
server := &http.Server{
    Addr: "127.0.0.1:80",
    Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Alo?"))
    })
}
go server.ListenAndServe()
client := ara.NewClient(ara.NewCustomResolver(map[string][]string{"example.com": {"127.0.0.1"}}))
res, _ := client.Get("http://example.com")
body, _ := ioutil.ReadAll(res.Body)
fmt.Println(string(body))
// Output: Alo?
```
