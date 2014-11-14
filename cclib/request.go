package cclib

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Request contains the API request basic information
type Request struct {
	Email          string
	Password       string
	Token          *Token
	TokenSourceUrl string
	Version        string
	Cache          string
	Url            string
	SslCheck       bool
	CaCerts        *x509.CertPool
}

// New request creates a new api request having:
//
// * User email
//
// * User password
//
// * User token
//
// Returns a new request pointer
func NewRequest(email string, password string, url string, token *Token, tokenSourceUrl string) *Request {
	return &Request{
		email,
		password,
		token,
		tokenSourceUrl,
		VERSION,
		CACHE,
		url,
		SSL_CHECK,
		CA_CERTS}
}

// SetEmail sets email address to a request
func (request *Request) SetEmail(email string) {
	request.Email = email
}

// SetPassword sets a password to a request
func (request *Request) SetPassword(password string) {
	request.Password = password
}

// SetToken sets a token to a request
func (request *Request) SetToken(token *Token) {
	request.Token = token
}

// SetCache sets a cache to a request
func (request *Request) SetCache(cache string) {
	request.Cache = cache
}

// SetUrl sets a URL to a request
func (request *Request) SetUrl(url string) {
	request.Url = url
}

// EnableSSLCheck enables the SSL certificate verification
func (request *Request) EnableSSLCheck() {
	request.SslCheck = true
}

// DisableSSLCheck disables the SSL certificate verification
func (request *Request) DisableSSLCheck() {
	request.SslCheck = false
}

// SetCaCerts sets a set of root CA to a request
func (request *Request) SetCaCerts(caCerts *x509.CertPool) {
	request.CaCerts = caCerts
}

// Post makes a POST request
func (request Request) Post(resource string, data url.Values) ([]byte, error) {
	return request.do(resource, "POST", data, false)
}

// Get makes a GET request
func (request Request) Get(resource string) ([]byte, error) {
	return request.do(resource, "GET", url.Values{}, false)
}

// Put makes a PUT request
func (request Request) Put(resource string, data url.Values) ([]byte, error) {
	return request.do(resource, "PUT", data, false)
}

// Delete makes a DELETE request
func (request Request) Delete(resource string) ([]byte, error) {
	return request.do(resource, "DELETE", url.Values{}, false)
}

// PostToken makes a POST request to the token source URL
func (request Request) PostToken() ([]byte, error) {
	return request.do("", "POST", nil, true)
}

func (request Request) do(resource string, method string, data url.Values, isTokenReq bool) ([]byte, error) {
	request_url := request.Url
	if isTokenReq {
		request_url = request.TokenSourceUrl
	}

	u, err := url.ParseRequestURI(request_url)
	if err != nil {
		return nil, err
	}

	if resource != "" {
		u.Path = resource
	}

	urlStr := fmt.Sprintf("%v", u)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !request.SslCheck,
			RootCAs:            request.CaCerts},
	}
	client := &http.Client{Transport: tr}

	r, err := http.NewRequest(method, urlStr, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	if request.Token != nil {
		r.Header.Add("Authorization", "cc_auth_token=\""+request.Token.Key()+"\"")
	} else if request.Email != "" && request.Password != "" {
		r.SetBasicAuth(request.Email, request.Password)
	}
	r.Header.Add("Host", u.Host)
	r.Header.Add("User-Agent", "gocclib/"+Version())
	if m := strings.ToUpper(method); m == "POST" || m == "PUT" {
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	r.Header.Add("Accept-Encoding", "compress, gzip")

	if DEBUG {
		fmt.Printf("DEBUG Request >>> %v\n", r)
	}

	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("DEBUG Request Error >>> %v\n", err)
		return nil, err
	}

	if err = checkResponse(resp); err != nil {
		if DEBUG {
			fmt.Printf("DEBUG Request Error >>> %v\n", err)
		}
		return nil, err
	}

	defer resp.Body.Close()
	if DEBUG {
		fmt.Printf("DEBUG Response >>> %v\n", resp)
		fmt.Printf("DEBUG Body >>> %v\n", resp.Body)
	}

	return ioutil.ReadAll(resp.Body)
}
