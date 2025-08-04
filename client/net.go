package client

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// Recibo el verbo (http.MethdodGet por ejemplo devuevle el string "GET"), reqURL es el path (/user por ej),
// El body se PSAA COMO INTERFACE (osea un struct)
func (rb *RequestBuilder) doRequest(verb string, reqURL string, reqBody interface{}) (result *Response) {

	start := time.Now()

	//Unimos base + path
	reqURL = rb.BaseURL + reqURL

	//Genramos un response
	result = new(Response)

	//Metodo que se encarga de hacer conversion del struct a un arrau de bytes
	body, err := rb.marshalReqBody(reqBody)
	if err != nil {
		result.Err = err
		if rb.LogTime {
			elapsed := time.Since(start)
			fmt.Printf("%s %s - time: %s\n", verb, reqURL, elapsed)
		}
		return result
	}

	//Mock es de test
	mock := getMock(verb, reqURL)

	var httpResp *http.Response
	if mock != nil {
		if mock.Err != nil {
			result.Err = mock.Err
			return result
		}
		httpResp = &http.Response{
			StatusCode: mock.RespHTTPCode,
			Header:     mock.RespHeaders,
			Body:       nopCloser{bytes.NewBufferString(mock.RespBody)},
		}
	} else {
		//Genera un client de http estandar de go
		client := rb.getClient()

		//Defino NewRequest con verb, url y body
		request, err := http.NewRequest(verb, reqURL, bytes.NewBuffer(body))
		if err != nil {
			result.Err = err
			if rb.LogTime {
				elapsed := time.Since(start)
				fmt.Printf("%s %s - time: %s\n", verb, reqURL, elapsed)
			}
			return result
		}

		// Set extra parameters, parametros ADICIONALES, se setean headers como el content type, accept, etc (si es json, xml, segun los campos del struct)
		rb.setParams(request)

		// Make the request
		httpResp, err = client.Do(request)
		if err != nil {
			result.Err = err
			if rb.LogTime {
				elapsed := time.Since(start)
				fmt.Printf("%s %s - time: %s\n", verb, reqURL, elapsed)
			}
			return result
		}
	}

	// Read response, leo respuesta
	defer httpResp.Body.Close()
	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		result.Err = err
		if rb.LogTime {
			elapsed := time.Since(start)
			fmt.Printf("%s %s - time: %s\n", verb, reqURL, elapsed)
		}
		return result
	}

	result.Response = httpResp
	result.byteBody = respBody

	if rb.LogTime {
		elapsed := time.Since(start)
		fmt.Printf("%s %s - time: %s\n", verb, reqURL, elapsed)
	}
	return result
}

func (rb *RequestBuilder) marshalReqBody(body interface{}) (b []byte, err error) {
	if body != nil {
		//ContentType es un enum de INT, siendo 0 (por defcto) JSON
		switch rb.ContentType {
		case JSON:
			b, err = json.Marshal(body)
		case XML:
			b, err = xml.Marshal(body)
		case BYTES:
			var ok bool
			b, ok = body.([]byte)
			if !ok {
				err = fmt.Errorf("bytes: body is %T(%v) not a byte slice", body, body)
			}
		}
	}
	return
}

func (rb *RequestBuilder) getClient() *http.Client {

	defaultTransport := &http.Transport{
		//MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
		//Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{Timeout: rb.getConnectionTimeout()}).DialContext,
		//ResponseHeaderTimeout: rb.getRequestTimeout(),
	}

	tr := defaultTransport

	rb.Client = &http.Client{Transport: tr}

	return rb.Client
}

/* func (rb *RequestBuilder) getRequestTimeout() time.Duration {
	switch {
	case rb.DisableTimeout:
		return 0
	case rb.Timeout > 0:
		return rb.Timeout
	default:
		return DefaultTimeout
	}
} */

func (rb *RequestBuilder) getConnectionTimeout() time.Duration {

	switch {
	case rb.DisableTimeout:
		return 0
	case rb.ConnectTimeout > 0:
		return rb.ConnectTimeout
	default:
		return DefaultConnectTimeout
	}
}

func (rb *RequestBuilder) setParams(req *http.Request) {

	// Default headers
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")

	// Encoding
	var cType string

	switch rb.ContentType {
	case JSON:
		cType = "json"
	case XML:
		cType = "xml"
	}

	if cType != "" {
		req.Header.Set("Accept", "application/"+cType)
		req.Header.Set("Content-Type", "application/"+cType)
	}

	for key, value := range rb.Headers {
		req.Header[key] = value
	}

}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func getMock(verb, reqURL string) *Mock {

	if len(mockMap) < 1 {
		return nil
	}

	if mock := mockMap[verb+" "+reqURL]; mock != nil {
		return mock
	}

	return nil
}
