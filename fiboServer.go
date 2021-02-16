package main

import (
	//	"sandbox"

	"encoding/json"
	//"fibo"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type serverConf struct {
	Port           int
	Addr           string
	CertPath       string
	PrivateKeyPath string
}

type cnf interface {
	getBaseServerUri() string
}

type getResponse struct {
	Version     string
	Description string
}

type getFiboResponse struct {
	Result              *[]uint64
	ComputationDuration int64
}

func assignAdd(res *uint64, op1 *uint64, op2 *uint64) {
	*res = *op1 + *op2
}

func Fibonaci2(input uint64, res *[]uint64) {
	switch input {
	case 0:
		(*res)[input] = 0
	case 1:
		(*res)[input] = 1
	default:
		defer assignAdd(&((*res)[input]), &((*res)[input-1]), &((*res)[input-2]))
		Fibonaci2(input-1, res)
		//fibonaci2(input-2, res)
	}
}

func (s *serverConf) getBaseServerUri() string {
	return s.Addr + ":" + strconv.FormatInt(int64(s.Port), 10)
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respBody := &getResponse{Version: "0.0.1", Description: "Fibonachi computation server abc"}
		respData, _ := (json.Marshal(respBody))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(respData)
	} else {
		w.WriteHeader(500)
		w.Write([]byte{})
		//fibo.Fibonaci2(1)
	}
}

func handleCompute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var respData []byte
		fmt.Printf("%s", r.URL.Path)
		splited := strings.Split(r.URL.Path, "/")
		if len(splited) != 3 {

		} else {
			value, _ := (strconv.Atoi(splited[2]))
			start := time.Now()

			result2 := make([]uint64, int(value+1))

			Fibonaci2(uint64(value), &result2)
			respBody := &getFiboResponse{Result: &result2, ComputationDuration: (time.Now().UnixNano() - start.UnixNano())}
			respData, _ = (json.Marshal(respBody))
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(respData)
	} else {

	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	data := make([]byte, 300)
	file, err := os.Open("./ServerConfig.json")
	count, err := file.Read(data)
	if err != nil {
		fmt.Printf("Failed to read config file reason: %s", err.Error())
		return
	}
	if count > 0 {
		var conf serverConf
		data := data[:count]
		fmt.Printf("Data: %s", data)
		err := json.Unmarshal(data, &conf)
		if err != nil {
			fmt.Printf("Failed to parse configuration: %s", err.Error())
			return
		}

		fmt.Printf("Server addr : %s", conf.getBaseServerUri())
		http2Srv := &http2.Server{}
		srv := &http.Server{Addr: conf.getBaseServerUri(), Handler: h2c.NewHandler(nil, http2Srv)}
		http.Handle("/version", http.HandlerFunc(handleVersion))
		http.Handle("/compute/", http.HandlerFunc(handleCompute))
		//http2.ConfigureServer(&srv, &http2Srv)
		//ConfigureServer(&srv, nil)
		go func(srvr *http.Server) {
			sig := <-sigs
			fmt.Println()
			fmt.Println(sig)
			srvr.Close()
		}(srv)

		fmt.Println("Starting server!")
		servErr := srv.ListenAndServe()
		if servErr != nil && servErr != http.ErrServerClosed {
			fmt.Printf("Failed to start server : %s", servErr.Error())
		}
	}

}
