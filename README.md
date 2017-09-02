# rawhttp

rawhttp is a [Go](https://golang.org/) package for making HTTP requests.
It intends to fill a nice that [https://golang.org/pkg/net/http/](net/http) does not cover:
having *complete* control over the requests being sent to the server.

Full documentation can be found on [GoDoc](https://godoc.org/github.com/tomnomnom/rawhttp).

## Example

```go
req, err := rawhttp.FromURL("POST", "https://httpbin.org")
if err != nil {
	log.Fatal(err)
}

// automatically set the host header
req.AutoSetHost()

req.Method = "PUT"
req.Hostname = "httpbin.org"
req.Port = "443"
req.Path = "/anything"
req.Query = "one=1&two=2"
req.Fragment = "anchor"
req.Proto = "HTTP/1.1"

req.AddHeader("Content-Type: application/x-www-form-urlencoded")

req.Body = "username=AzureDiamond&password=hunter2"

// automatically set the Content-Length header
req.AutoSetContentLength()

fmt.Printf("%s\n\n", req.String())

resp, err := rawhttp.Do(req)
if err != nil {
	log.Fatal(err)
}

for _, h := range resp.Headers() {
	fmt.Printf("< %s\n", h)
}

fmt.Printf("\n%s\n", resp.Body())
```

```
PUT /anything?one=1&two=2#anchor HTTP/1.1
Host: httpbin.org
Content-Type: application/x-www-form-urlencoded
Content-Length: 38

username=AzureDiamond&password=hunter2

< Connection: keep-alive
< Server: meinheld/0.6.1
< Date: Sat, 02 Sep 2017 13:22:06 GMT
< Content-Type: application/json
< Access-Control-Allow-Origin: *
< Access-Control-Allow-Credentials: true
< X-Powered-By: Flask
< X-Processed-Time: 0.000869989395142
< Content-Length: 443
< Via: 1.1 vegur

{
  "args": {
    "one": "1",
    "two": "2"
  },
  "data": "",
  "files": {},
  "form": {
    "password": "hunter2",
    "username": "AzureDiamond"
  },
  "headers": {
    "Connection": "close",
    "Content-Length": "38",
    "Content-Type": "application/x-www-form-urlencoded",
    "Host": "httpbin.org"
  },
  "json": null,
  "method": "PUT",
  "origin": "123.123.123.123",
  "url": "https://httpbin.org/anything?one=1&two=2"
}
```
