package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

type EmptyRequest struct {
}

func getServerVersion() {

	const url = "http://127.0.0.1:80/compute/100"
	client := http.Client{Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		}}}
	/*response, err := client.Get( url)
	if err != nil {
		fmt.Printf("Request failed because: %s", err)
		return
	}*/
	jsonReq, _ := json.Marshal(EmptyRequest{})
	response, err := client.Post(url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		fmt.Printf("Request failed because: %s", err)
		return
	}
	respBuf := make([]byte, 1024)
	n, _ := response.Body.Read(respBuf)
	respBuf = respBuf[:n]
	fmt.Println(string(respBuf))
}

func main() {
	getServerVersion()
}
