package main

import (
	//	"sandbox"

	"Projects/fibo"
	"Projects/myServer"
	"encoding/json"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

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

const ( // iota is reset to 0
	RESP_SUCCESS           = 0
	RESP_400_INVALID_PARAM = 4001
	RESP_400_WRONG_METHOD  = 4002
	RESP_400_INVALID_URI   = 4003
	UNKNOWN                = 5000
)

type errorResponse struct {
	ErrorDescription string
	errCode          uint
	errSpec          string
}

func assignAdd(res *uint64, op1 *uint64, op2 *uint64) {
	*res = *op1 + *op2
}

/*func Fibonaci2(input uint64, res *[]uint64) {
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
}*/

func handleVersion(w *http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respBody := &getResponse{Version: "0.0.1", Description: "Fibonachi computation server abc"}
		respData, _ := (json.Marshal(respBody))
		(*w).Header().Add("Content-Type", "application/json")
		(*w).WriteHeader(200)
		(*w).Write(respData)
	} else {
		(*w).WriteHeader(500)
		(*w).Write([]byte{})
		//fibo.Fibonaci2(1)
	}
}

func (e *errorResponse) ConstructErrorDescription() {
	switch e.errCode {
	case RESP_SUCCESS:
		return
	case RESP_400_INVALID_PARAM:
		e.ErrorDescription = "Following problem occured while procesing your parameter to fibonaci: " + e.errSpec
	case RESP_400_WRONG_METHOD:
		e.ErrorDescription = "This server does not support used method"
	case RESP_400_INVALID_URI:
		e.ErrorDescription = "Uri used for GET request has to be in format /compute/X where X is number greater than 0"
	default:
		e.ErrorDescription = "Unknown error occured during computation!"
	}
}

func (e *errorResponse) getHttpRespCode() (retval uint) {
	retval = e.errCode / 10
	return
}

func handleCompute(w *http.ResponseWriter, r *http.Request) {
	var respData []byte
	errCtx := errorResponse{errCode: RESP_SUCCESS}

	fmt.Printf("%s", r.URL.Path)
	splited := strings.Split(r.URL.Path, "/")
	if len(splited) != 3 {
		errCtx.errCode = RESP_400_INVALID_URI
	} else {
		value, err := (strconv.Atoi(splited[2]))
		if err != nil {
			errCtx.errCode = RESP_400_INVALID_PARAM
			errCtx.errSpec = err.Error()
		} else {
			start := time.Now()

			result2 := make([]uint64, value+1)

			fibo.Fibonaci2(uint64(value), &result2)
			respData, _ = (json.Marshal(&getFiboResponse{Result: &result2, ComputationDuration: (time.Now().UnixNano() - start.UnixNano())}))
		}
	}

	if errCtx.errCode == 0 {
		(*w).Header().Add("Content-Type", "application/json")
	} else {
		(*w).Header().Add("Content-Type", "application/problem+json")
		errCtx.ConstructErrorDescription()
		respData, _ = json.Marshal(errCtx)
	}
	(*w).Write(respData)
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	h2cServer := &myServer.SimpleServer{}

	h2cServer.GetMux().Add("/version", "GET", handleVersion)
	h2cServer.GetMux().Add("/compute/", "GET", handleCompute)

	h2cServer.Init("./serverConfig.json")
	//http2.ConfigureServer(&srv, &http2Srv)
	//ConfigureServer(&srv, nil)
	go func(srvr *myServer.SimpleServer) {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		err := srvr.Stop()
		if err != nil {
			panic(err)
		}
	}(h2cServer)
	servErr := h2cServer.Start()
	if servErr != nil {
		fmt.Printf(servErr.Error())
	}
}
