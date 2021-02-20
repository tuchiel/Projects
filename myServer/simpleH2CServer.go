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
	mux  SimpleServerMux
}

const default_handler_tag string = "__defaultHandlerTag"

func createErrorFromError(description string, err error) error {
	return errors.New(description + err.Error())
}

type HandlerFunc func(*http.ResponseWriter, *http.Request)
type MethodMux map[string]HandlerFunc
type SimpleServerMux map[string]MethodMux

type HandlerWrapper struct {
	m MethodMux
}

func (h HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, exists := h.m[r.Method]
	if !exists {
		handler, exists = h.m[default_handler_tag]
	}
	handler(&w, r)
}

func (mux *SimpleServerMux) Add(path string, method string, handler HandlerFunc) error {
	switch {
	case method == http.MethodGet,
		method == http.MethodPost,
		method == http.MethodPut,
		method == http.MethodPatch,
		method == http.MethodConnect,
		method == http.MethodHead,
		method == http.MethodOptions,
		method == http.MethodTrace:
	default:
		return errors.New("specified method is not allowed")
	}
	if (*mux)[path] == nil {
		(*mux)[path] = make(map[string]HandlerFunc)
	}
	(*mux)[path][method] = handler
	return nil
}

func (mux *SimpleServerMux) AddDefault(path *string, handler HandlerFunc) {
	if (*mux)[*path] == nil {
		(*mux)[*path] = make(map[string]HandlerFunc)
	}
	(*mux)[*path][default_handler_tag] = handler
}

func (server *SimpleServer) GetMux() *SimpleServerMux {
	if server.mux == nil {
		server.mux = make(SimpleServerMux)
	}
	return &server.mux
}

func defaultServerHandler(w *http.ResponseWriter, r *http.Request) {
	(*w).WriteHeader(405)
	(*w).Write([]byte{})
}

func (server *SimpleServer) Init(configFilePath string) (err error) {
	const (
		bufferSize = 10
	)
	var (
		file  *os.File
		count int
	)
	file, err = os.Open(configFilePath)
	if err != nil {
		err = createErrorFromError("Failed to open server config file reason:", err)
		return
	}
	buf := make([]byte, bufferSize)
	data := make([]byte, 0)
	for i := 0; ; i++ {
		count, err = file.Read(buf)
		if err != nil {
			err = createErrorFromError("Failed to read server config file reason:", err)
			return
		}
		data = append(data, buf...)
		if count < bufferSize {
			data = data[:i*bufferSize+count]
			break
		}
	}
	fmt.Printf("Server configuration: %s", data)
	err = json.Unmarshal(data, &server.conf)
	if err != nil {
		err = createErrorFromError("Failed to parse configuration reason:", err)
		return
	}

	for _, handlers := range server.mux {
		_, hasDefaultHandler := handlers[default_handler_tag]
		if !hasDefaultHandler {
			handlers[default_handler_tag] = defaultServerHandler
		}
	}

	fmt.Printf("Server addr : %s", server.conf.getBaseServerUri())
	server.srv = &http.Server{Addr: server.conf.getBaseServerUri(), Handler: h2c.NewHandler(http.DefaultServeMux, &http2.Server{})}
	for uri, uriHandlers := range server.mux {
		http.Handle(uri, HandlerWrapper{uriHandlers})
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
