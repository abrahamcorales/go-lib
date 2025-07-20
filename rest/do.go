package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Do perform the http request taking in consideration the fields of the request
// return the response
func (r *Request) Do() *Response {
	var ctx context.Context
	var cancel context.CancelFunc
	if r.isMocked {
		return handleMockRequest(r, r.URL, r.Method)
	}

	url := r.client.Config.BaseURL + r.URL

	if r.cached {
		log.Printf("Checking cache for url: %s", url)
		_, cachedResponse := r.client.Cache.Get(r.ctx, r.Method+url)
		if cachedResponse != nil {
			return cachedResponse.(*Response)
		}
	}

	if r.TimeoutInMillis > 0 {
		time := time.Duration(r.TimeoutInMillis) * time.Millisecond
		ctx, cancel = context.WithTimeout(r.ctx, time)
		defer cancel()
	} else {
		ctx = r.ctx

	}

	var req *http.Request
	var err error

	if r.Body != nil {
		var b []byte
		b, err = json.Marshal(r.Body)
		if err != nil {
			return &Response{
				StatusCode: 800,
				Error:      err,
			}
		}
		req, err = http.NewRequest(r.Method, url, bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(r.Method, url, nil)
	}

	if r.ctx != nil {
		req = req.WithContext(ctx)
	}

	if err != nil {
		return nil
	}

	if r.Headers != nil {
		for k, v := range r.Headers {
			headersConcat := strings.Join(v, ",")
			req.Header.Set(k, headersConcat)
		}
	}

	if r.AuthorizationToken != nil {
		req.Header.Set("Authorization", "Bearer "+*r.AuthorizationToken)
	}

	start := time.Now()
	res, err := r.client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	defer func(r *http.Response) {
		if r != nil && r.Body != nil {
			r.Body.Close()
		}
	}(res)

	if err != nil {
		statusCode := 500
		if res != nil {
			statusCode = res.StatusCode
		}
		return &Response{
			StatusCode: statusCode,
			Duration:   elapsed,
			Error:      errors.Wrap(err, "error on url: "+url),
		}
	}

	bodyBytes := []byte{}
	if res.Body != nil {
		bodyBytes, err = io.ReadAll(res.Body)
		if err != nil {
			return &Response{
				StatusCode: 800,
				Error:      errors.Wrap(err, "error reading response body"),
			}
		}
	}

	response := &Response{
		StatusCode:  res.StatusCode,
		Headers:     res.Header,
		RawResponse: res,
		//Body:        res.Body,
		BodyBytes: bodyBytes,
		Error:     nil,
		Duration:  elapsed,
	}

	if r.cached {
		log.Printf("Caching response for url: %s with ttl: %v", url, r.cacheTTL)
		r.client.Cache.SaveWithTTL(r.ctx, r.Method+url, response, r.cacheTTL)
	}

	//create a response object
	return response
}
