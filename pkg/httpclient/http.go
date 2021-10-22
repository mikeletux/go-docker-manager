package httpclient

type HttpResponse struct {
	// Code is the response code from the HTTP request
	StatusCode int
	// Body is the returned body from the HTTP query
	Body []byte
}

type HttpClient interface {
	// Get performs a HTTP GET method agains an urlEndpoint using HTTP headers. It returns an HttpResponse
	Get(urlEndpoint string, headers map[string]string) (*HttpResponse, error)

	// Post performs a HTTP POST method agains an urlEndpoint using HTTP headers and a body. It returns an HttpResponse
	Post(urlEndpoint string, headers map[string]string, body string) (*HttpResponse, error)

	// Delete performs a HTTP DELETE method agains an urlEndpoint using HTTP headers. It returns an HttpResponse
	Delete(urlEndpoint string, headers map[string]string) (*HttpResponse, error)
}
