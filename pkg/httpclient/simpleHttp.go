package httpclient

import (
	"io"
	"net/http"
	"strings"
)

type SimpleHttpClient struct {
	HttpClient *http.Client
}

func NewSimpleHttpClient() *SimpleHttpClient {

	return &SimpleHttpClient{HttpClient: &http.Client{}}
}

// Get performs a HTTP GET method agains an urlEndpoint using HTTP headers. It returns an HttpResponse
func (s *SimpleHttpClient) Get(urlEndpoint string, headers map[string]string) (*HttpResponse, error) {
	return s.runRequest(urlEndpoint, "GET", headers, "")
}

// Post performs a HTTP POST method agains an urlEndpoint using HTTP headers and a body. It returns an HttpResponse
func (s *SimpleHttpClient) Post(urlEndpoint string, headers map[string]string, body string) (*HttpResponse, error) {
	return s.runRequest(urlEndpoint, "POST", headers, body)
}

// Delete performs a HTTP DELETE method agains an urlEndpoint using HTTP headers. It returns an HttpResponse
func (s *SimpleHttpClient) Delete(urlEndpoint string, headers map[string]string) (*HttpResponse, error) {
	return s.runRequest(urlEndpoint, "DELETE", headers, "")
}

func (s *SimpleHttpClient) runRequest(urlEndpoint string, method string, headers map[string]string, body string) (*HttpResponse, error) {
	// Create the HTTP Request
	req, err := http.NewRequest(method, urlEndpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Add the headers to the query
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Perform the query
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read all resp body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{StatusCode: resp.StatusCode,
		Body: respBody}, nil
}
