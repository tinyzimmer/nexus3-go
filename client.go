package nexus

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// String is a convenience function for converting strings to pointers
// for use in inputs.
func String(nonPtr string) *string {
	return &nonPtr
}

// Bool is a convenience function for returning the pointer to a boolean value.
func Bool(nonPtr bool) *bool {
	return &nonPtr
}

// Int is a convenience function for returning the pointer to an integer.
func Int(nonPtr int) *int {
	return &nonPtr
}

// Nexus represents the main interface for interacting with Nexus.
//
// Creating a Client
//
// Create a client and panic on unreachable server or bad credentials
//
//     client, err := nexus.New("http://localhost:8081", "username", "password")
//     if err != nil {
//         panic(err)
//     }
//
type Nexus struct {
	client   *http.Client
	host     string
	username string
	password string
}

// New creates a Nexus client with the given parameters
func New(host string, username string, password string) (n *Nexus, err error) {
	n = &Nexus{}
	n.client = &http.Client{}
	n.host = host
	n.username = username
	n.password = password
	err = n.Status()
	return
}

// NewRequest returns an HTTP request for the given method, endpoint, and body
// then sets the basic authentication on the request.
func (n *Nexus) NewRequest(method string, endpoint string, args map[string]string, body []byte, contentType string) (req *http.Request, err error) {
	if contentType == "" {
		contentType = "application/json"
	}
	u, err := url.Parse(fmt.Sprintf("%s/%s", n.host, endpoint))
	if err != nil {
		return
	}
	if args != nil {
		u = n.BuildQueryURL(u, args)
	}
	if body == nil {
		req, err = http.NewRequest(method, u.String(), nil)
	} else {
		req, err = http.NewRequest(method, u.String(), bytes.NewBuffer(body))
	}
	if err != nil {
		return
	}
	req.SetBasicAuth(n.username, n.password)
	req.Header.Set("Content-Type", contentType)
	return
}

// BuildQueryURL returns a URL formatted with query parameters for a GET request
func (n *Nexus) BuildQueryURL(rawURL *url.URL, args map[string]string) (u *url.URL) {
	q := rawURL.Query()
	for k, v := range args {
		q.Set(k, v)
	}
	rawURL.RawQuery = q.Encode()
	u = rawURL
	return
}

// Do preforms an HTTP request of the pre-packaged request object and returns
// the body or any errors. If provided, an error will be created with with the text
// of the cooresponding status code in the `statusMap`.
func (n *Nexus) Do(req *http.Request, statusMap map[int]string, resToErr bool) (body []byte, err error) {
	resp, err := n.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		if resToErr {
			content, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(string(content))
			return
		}
		if statusMap != nil {
			if status, ok := statusMap[resp.StatusCode]; ok {
				err = fmt.Errorf(status)
				return
			}
		}
		// Safety belt
		err = fmt.Errorf("%s %s returned a status code of %v", req.Method, req.URL.String(), resp.StatusCode)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	return
}

// Status is used as a "ping" of the server. The endpoint returns a non-200
// code when the server is unable to serve requests or the credentials are invalid.
func (n *Nexus) Status() (err error) {
	req, err := n.NewRequest("GET", "service/rest/v1/status", nil, nil, "")
	if err != nil {
		return
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("Credentials are invalid or Nexus is unable to serve requests")
	}
	return
}
