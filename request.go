package rawhttp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// A Requester defines the bare minimum set of methods needed to make an HTTP request.
type Requester interface {
	// IsTLS should return true if the connection should be made using TLS
	IsTLS() bool

	// Host should return a hostname:port pair
	Host() string

	// String should return the request as a string E.g:
	//   GET / HTTP/1.1\r\nHost:...
	String() string

	// GetTimeout returns the timeout for a request
	GetTimeout() time.Duration
}

type CusRequester struct {
	IsTLS bool
	Host string
	String string
	Timeout time.Duration
}

// Request is the main implementation of Requester. It gives you
// fine-grained control over just about everything to do with the
// request, but with the posibility of sane defaults.
type Request struct {
	// TLS should be true if TLS should be used
	TLS bool

	// Method is the HTTP verb. E.g. GET
	Method string

	// Scheme is the protocol scheme. E.g. https
	Scheme string

	// Hostname is the hostname to connect to. E.g. localhost
	Hostname string

	// Port is the port to connect to. E.g. 80
	Port string

	// Path is the path to request. E.g. /security.txt
	Path string

	// Query is the query string of the path. E.g. q=searchterm&page=3
	Query string

	// Fragment is the bit after the '#'. E.g. pagesection
	Fragment string

	// Proto is the protocol specifier in the first line of the request.
	// E.g. HTTP/1.1
	Proto string

	// Headers is a slice of headers to send. E.g:
	//   []string{"Host: localhost", "Accept: text/plain"}
	Headers []string

	// Body is the 'POST' data to send. E.g:
	//   username=AzureDiamond&password=hunter2
	Body string

	// EOL is the string that should be used for line endings. E.g. \r\n
	EOL string

	// Deadline
	Timeout time.Duration
}

// FromURL returns a *Request for a given method and URL and any
// error that occured during parsing the URL. Sane defaults are
// set for all of *Request's fields.
func FromURL(method, rawurl string) (*Request, error) {
	r := &Request{}

	u, err := url.Parse(rawurl)
	if err != nil {
		return r, err
	}

	r.TLS = u.Scheme == "https"
	r.Method = method
	r.Scheme = u.Scheme
	r.Hostname = u.Hostname()
	r.Port = u.Port()
	r.Path = u.Path
	if u.RawPath != ""{
		r.Path = u.RawPath
	}
	r.Query = u.RawQuery
	r.Fragment = u.Fragment
	if u.RawFragment != ""{
		r.Fragment = u.RawFragment
	}
	r.Proto = "HTTP/1.1"
	r.EOL = "\r\n"
	r.Timeout = time.Second * 30

	if r.Path == "" {
		r.Path = "/"
	}

	if r.Port == "" {
		if r.TLS {
			r.Port = "443"
		} else {
			r.Port = "80"
		}
	}

	return r, nil

}

// IsTLS returns true if TLS should be used
func (r Request) IsTLS() bool {
	return r.TLS
}

// Host returns the hostname:port pair to connect to
func (r Request) Host() string {
	return r.Hostname + ":" + r.Port
}

// AddHeader adds a header to the *Request
func (r *Request) AddHeader(h string) {
	r.Headers = append(r.Headers, h)
}

// Header finds and returns the value of a header on the request.
// An empty string is returned if no match is found.
func (r Request) Header(search string) string {
	search = strings.ToLower(search)

	for _, header := range r.Headers {

		p := strings.SplitN(header, ":", 2)
		if len(p) != 2 {
			continue
		}

		if strings.ToLower(p[0]) == search {
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}

// AutoSetHost adds a Host header to the request
// using the value of Request.Hostname
func (r *Request) AutoSetHost() {
	r.AddHeader(fmt.Sprintf("Host: %s", r.Hostname))
}

// AutoSetContentLength adds a Content-Length header
// to the request with the length of Request.Body as the value
func (r *Request) AutoSetContentLength() {
	r.AddHeader(fmt.Sprintf("Content-Length: %d", len(r.Body)))
}

// fullPath returns the path including query string and fragment
func (r Request) fullPath() string {
	q := ""
	if r.Query != "" {
		q = "?" + r.Query
	}

	f := ""
	if r.Fragment != "" {
		f = "#" + r.Fragment
	}
	return r.Path + q + f
}

// URL forms and returns a complete URL for the request
func (r Request) URL() string {
	return fmt.Sprintf(
		"%s://%s%s",
		r.Scheme,
		r.Host(),
		r.fullPath(),
	)
}

// RequestLine returns the request line. E.g. GET / HTTP/1.1
func (r Request) RequestLine() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.fullPath(), r.Proto)
}

// String returns a plain-text version of the request to be sent to the server
func (r Request) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("%s%s", r.RequestLine(), r.EOL))

	for _, h := range r.Headers {
		b.WriteString(fmt.Sprintf("%s%s", h, r.EOL))
	}

	b.WriteString(r.EOL)

	b.WriteString(r.Body)

	return b.String()
}

// GetTimeout returns the timeout for a request
func (r Request) GetTimeout() time.Duration {
	// default 30 seconds
	if r.Timeout == 0 {
		return time.Second * 30
	}
	return r.Timeout
}

// RawRequest is the most basic implementation of Requester. You should
// probably only use it if you're doing something *really* weird
type RawRequest struct {
	// TLS should be true if TLS should be used
	TLS bool

	// Hostname is the name of the host to connect to. E.g: localhost
	Hostname string

	// Port is the port to connect to. E.g.: 80
	Port string

	// Request is the actual message to send to the server. E.g:
	//   GET / HTTP/1.1\r\nHost:...
	Request string

	// Timeout for the request
	Timeout time.Duration
}

// IsTLS returns true if the connection should use TLS
func (r RawRequest) IsTLS() bool {
	return r.TLS
}

// Host returns the hostname:port pair
func (r RawRequest) Host() string {
	return r.Hostname + ":" + r.Port
}

// String returns the message to send to the server
func (r RawRequest) String() string {
	return r.Request
}

// GetTimeout returns the timeout for the request
func (r RawRequest) GetTimeout() time.Duration {
	// default 30 seconds
	if r.Timeout == 0 {
		return time.Second * 30
	}
	return r.Timeout
}

// Do performs the HTTP request for the given Requester and returns
// a *Response and any error that occured
func Do(req Requester) (*Response, error) {
	var conn net.Conn
	var connerr error

	// This needs timeouts because it's fairly likely
	// that something will go wrong :)
	if req.IsTLS() {
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		// This library is meant for doing stupid stuff, so skipping cert
		// verification is actually the right thing to do
		conf := &tls.Config{RootCAs: roots, InsecureSkipVerify: true}

		conn, connerr = tls.DialWithDialer(&net.Dialer{
			Timeout: req.GetTimeout(),
		}, "tcp", req.Host(), conf)
	} else {
		conn, connerr = net.DialTimeout("tcp", req.Host(),req.GetTimeout())
	}

	if connerr != nil {
		return nil, connerr
	}

	fmt.Fprint(conn, req.String())
	fmt.Fprint(conn, "\r\n")

	conn.SetDeadline(time.Now().Add(req.GetTimeout()))
	defer conn.Close()

	return newResponse(conn)
}

func DoWithCus(req CusRequester) (*Response, error) {
	var conn net.Conn
	var connerr error

	// This needs timeouts because it's fairly likely
	// that something will go wrong :)
	if req.IsTLS {
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		// This library is meant for doing stupid stuff, so skipping cert
		// verification is actually the right thing to do
		conf := &tls.Config{RootCAs: roots, InsecureSkipVerify: true}

		conn, connerr = tls.DialWithDialer(&net.Dialer{
			Timeout: req.Timeout,
		}, "tcp", req.Host, conf)
	} else {
		conn, connerr = net.DialTimeout("tcp", req.Host,req.Timeout)
	}

	if connerr != nil {
		return nil, connerr
	}

	fmt.Fprint(conn, req.String)
	fmt.Fprint(conn, "\r\n")

	conn.SetDeadline(time.Now().Add(req.Timeout))
	defer conn.Close()

	return newResponse(conn)
}

func DoWithSni(sni string,req Requester) (*Response, error) {
	var conn net.Conn
	var connerr error

	// This needs timeouts because it's fairly likely
	// that something will go wrong :)
	if req.IsTLS() {
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		// This library is meant for doing stupid stuff, so skipping cert
		// verification is actually the right thing to do
		conf := &tls.Config{RootCAs: roots, InsecureSkipVerify: true,ServerName: sni}

		conn, connerr = tls.DialWithDialer(&net.Dialer{
			Timeout: req.GetTimeout(),
		}, "tcp", req.Host(), conf)
	} else {
		conn, connerr = net.DialTimeout("tcp", req.Host(),req.GetTimeout())
	}

	if connerr != nil {
		return nil, connerr
	}

	fmt.Fprint(conn, req.String())
	fmt.Fprint(conn, "\r\n")

	conn.SetDeadline(time.Now().Add(req.GetTimeout()))
	defer conn.Close()

	return newResponse(conn)
}
