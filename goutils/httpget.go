package goutils

import (
	"log"
	"net/http"
	"fmt"
	"time"
	"crypto/tls"
	"os"
	"io"

	"github.com/gobs/httpclient"
)

type HttpClient struct {
	Client	*httpclient.HttpClient
	Target 	string
}

// Get URL
func GetUrl(target string) {
	httpClient := NewClient(target)
	request := httpClient.Request("get")
	executedRequest := httpClient.executeRequest(request)
	respOK := ResponseHandler("outfile", executedRequest)
	if respOK {
	}
	return
}

func NewClient(target string) *HttpClient {
	httpClient := httpclient.NewHttpClient(target)
	httpClient.SetTimeout(time.Minute * time.Duration(60))
	return &HttpClient{
		Client: httpClient,
	}
}

// Format the request
func (httpClient *HttpClient) Request(method string) *http.Request {
	request := httpClient.Client.Request(method, httpClient.Target, nil, nil)
	return request
}

// Execute the formatted request
func (httpClient *HttpClient) executeRequest(request *http.Request) *httpclient.HttpResponse {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, err := httpClient.Client.Do(request)
	log.Println(fmt.Sprintf("message=request_sent url=%s", request.URL))
	// Process response
	if err != nil {
		log.Println(fmt.Sprintf("status=failed_client_request host=%s error=%s", request.URL, err))
		response.Body.Close()
		return nil
	}
	return response
}

// Handle the HTTP response and write incoming bytes to a file
func ResponseHandler(outFilename string, response *httpclient.HttpResponse) bool {
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Println(fmt.Sprintf("status=http_request_failed status=%d", response.StatusCode))
		return false
		}
	out, err := os.Create(outFilename)
	if err != nil {
		log.Println(fmt.Sprintf("status=out_file_create_fail file=%s", outFilename))
	}
	defer out.Close()
	io.Copy(out, response.Body)
	return true
}