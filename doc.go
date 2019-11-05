/*
Package ara provides a dialer with customizable resolver.

It can be used with http.Client and http.Transport to alter host lookups.
For example, with a custom resolver that maps google.com to 127.0.0.1,
you can get httpClient.Get("http://google.com") to connect the localhost.
*/
package ara
