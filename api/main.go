package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"reflect"

	"github.com/go-msvc/errors"
	"github.com/gorilla/mux"
	"github.com/stewelarend/logger"
)

func main() {
	addrPtr := flag.String("address", ":8080", "Server address")
	flag.Parse()
	r := mux.NewRouter()

	r.HandleFunc("/accounts", hdlr(accLst)).Methods(http.MethodGet)
	r.HandleFunc("/accounts", hdlr(accAdd)).Methods(http.MethodPost)
	r.HandleFunc("/accounts/{id}", hdlr(accGet)).Methods(http.MethodGet)
	// r.HandleFunc("/accounts/{id}", hdlr(accUpd)).Methods(http.MethodPut)
	// r.HandleFunc("/accounts/{id}", hdlr(accDel)).Methods(http.MethodDelete)

	r.HandleFunc("/transactions", hdlr(txLst)).Methods(http.MethodGet)
	r.HandleFunc("/transactions", hdlr(txAdd)).Methods(http.MethodPost)
	r.HandleFunc("/transactions/{id}", hdlr(txGet)).Methods(http.MethodGet)
	// r.HandleFunc("/transactions/{id}", hdlr(txUpd)).Methods(http.MethodPut)
	// r.HandleFunc("/transactions/{id}", hdlr(txDel)).Methods(http.MethodDelete)

	http.ListenAndServe(*addrPtr, r)
}

type CtxLogger struct{}

type Validator interface{ Validate() error }

func hdlr(fnc interface{}) http.HandlerFunc {
	fncType := reflect.TypeOf(fnc)
	reqType := fncType.In(1)
	//resType := fncType.Out(0)
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		ctx := context.Background()
		log := logger.New().WithLevel(logger.LevelDebug)
		ctx = context.WithValue(ctx, CtxLogger{}, log)
		log.Debugf("HTTP %s %s", httpReq.Method, httpReq.URL.Path)

		httpStatus := http.StatusInternalServerError
		var err error
		var res interface{}
		defer func() {
			if res != nil {
				if err = json.NewEncoder(httpRes).Encode(res); err != nil {
					err = errors.Wrapf(err, "failed to encode JSON response")
					httpStatus = http.StatusInternalServerError
				}
			}

			if httpStatus != http.StatusOK {
				if err == nil {
					err = errors.Errorf("failed without an error")
				}
				http.Error(httpRes, err.Error(), httpStatus)
				log.Errorf("HTTP %s %s: %+v", httpReq.Method, httpReq.URL.Path, err)
			}
		}()

		reqValuePtr := reflect.New(reqType)
		if httpReq.Body != nil {
			if err = json.NewDecoder(httpReq.Body).Decode(reqValuePtr.Interface()); err != nil && err != io.EOF {
				err = errors.Wrapf(err, "cannot decode JSON body")
				return
			}
			if validator, ok := reqValuePtr.Interface().(Validator); ok {
				if err = validator.Validate(); err != nil {
					httpStatus = http.StatusBadRequest
					return
				}
			}
			log.Debugf("req: (%T)%+v", reqValuePtr.Elem().Interface(), reqValuePtr.Elem().Interface())
		}

		resultValues := reflect.ValueOf(fnc).Call(
			[]reflect.Value{
				reflect.ValueOf(ctx),
				reqValuePtr.Elem(),
			},
		)
		if !resultValues[1].IsNil() {
			err = resultValues[1].Interface().(error)
			return
		}

		if !resultValues[0].IsNil() {
			res = resultValues[0].Elem().Interface()
			log.Debugf("%+v -> %+v", reqValuePtr.Elem().Interface(), res)
		}
		httpStatus = http.StatusOK
	}
}
