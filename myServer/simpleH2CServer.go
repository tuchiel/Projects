package myServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type serverConf struct {
	Port           int
	Addr           string
	CertPath       string
	PrivateKeyPath string
}

func (s *serverConf) getBaseServerUri() string {
	return s.Addr + ":" + strconv.FormatInt(int64(s.Port), 10)
}

type SimpleServer struct {
	conf serverConf
	srv  *http.Server
}

func createErrorFromError(description string, err error) error {
	return errors.New(description + err.Error())
}

func (server *SimpleServer) Init(configFilePath string, serverMux *map[string]http.HandlerFunc) (err error) {
	const (
		bufResize = 300
	)
	data := make([]byte, bufResize)
	var (
		file  *os.File
		count int
	)
	for i := 0; ; i++ {

		file, err = os.Open(configFilePath)
		if err != nil {
			err = createErrorFromError("Failed to open server config file reason:", err)
			return
		}
		count, err = file.Read(data)
		if err != nil {
			err = createErrorFromError("Failed to read server config file reason:", err)
			return
		}
		if count < bufResize {
			data = data[:i*bufResize+count]
			break
		} else {
			data = data[:bufResize]
		}
	}
	fmt.Printf("Server configuration: %s", data)
	err = json.Unmarshal(data, &server.conf)
	if err != nil {
		err = createErrorFromError("Failed to parse configuration reason:", err)
		return
	}

	fmt.Printf("Server addr : %s", server.conf.getBaseServerUri())
	server.srv = &http.Server{Addr: server.conf.getBaseServerUri(), Handler: h2c.NewHandler(nil, &http2.Server{})}
	for uri, handler := range *serverMux {
		http.Handle(uri, handler)
	}
	return nil
}

func (server *SimpleServer) Start() error {
	fmt.Println("Starting server!")
	servErr := server.srv.ListenAndServe()
	if servErr != http.ErrServerClosed {
		return createErrorFromError("Failed to start server reason :", servErr)
	} else {
		return nil
	}
}

func (server *SimpleServer) Stop() error {
	fmt.Println("Stopping server!")
	return server.srv.Close()
}
