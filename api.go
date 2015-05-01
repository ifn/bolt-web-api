package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

type Response struct {
	Error string `json:"error"`
}

type DataResponse struct {
	Response
	Data string `json:"data"`
}

var (
	ErrInternal = errors.New("Internal error")
)

type ErrHandlerFunc func(r *http.Request) error
type DataErrHandlerFunc func(r *http.Request) (string, error)

type ErrHandler struct {
	err     error
	handler ErrHandlerFunc
}

func (self *ErrHandler) Handle(r *http.Request) {
	defer func() {
		if exc := recover(); exc != nil {
			log.Printf("%s:\n%s", exc, debug.Stack())
			self.err = ErrInternal
		}
	}()

	self.err = self.handler(r)
}

func (self *ErrHandler) Response() interface{} {
	if self.err != nil {
		return Response{self.err.Error()}
	}
	return Response{}
}

type DataErrHandler struct {
	data    string
	err     error
	handler DataErrHandlerFunc
}

func (self *DataErrHandler) Handle(r *http.Request) {
	defer func() {
		if exc := recover(); exc != nil {
			log.Printf("%s:\n%s", exc, debug.Stack())
			self.err = ErrInternal
		}
	}()

	self.data, self.err = self.handler(r)
}

func (self *DataErrHandler) Response() interface{} {
	if self.err != nil {
		return DataResponse{Response{self.err.Error()}, ""}
	}
	return DataResponse{Data: self.data}
}

type handler interface {
	Handle(*http.Request)
	Response() interface{}
}

func jsonResp(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		h.Handle(r)

		encoder.Encode(h.Response())
	}
}

//

var (
	ErrNoBucket = errors.New("bucket doesn't exist")
)

func CreateBucketHandler(bs *BoltServer) *ErrHandler {
	return &ErrHandler{handler: func(r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucket([]byte(buck_name))
			return
		})
	}}
}

func CreateBucketIfNotExistsHandler(bs *BoltServer) *ErrHandler {
	return &ErrHandler{handler: func(r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucketIfNotExists([]byte(buck_name))
			return
		})
	}}
}

func DeleteBucketHandler(bs *BoltServer) *ErrHandler {
	return &ErrHandler{handler: func(r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) error {
			return tx.DeleteBucket([]byte(buck_name))
		})
	}}
}

//

func PutHandler(bs *BoltServer) *ErrHandler {
	return &ErrHandler{handler: func(r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]
		key := mux.Vars(r)["key"]

		val, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		return bs.db.Update(func(tx *bolt.Tx) error {
			if buck := tx.Bucket([]byte(buck_name)); buck != nil {
				return buck.Put([]byte(key), val)
			}
			return ErrNoBucket
		})
	}}
}

func GetHandler(bs *BoltServer) *DataErrHandler {
	return &DataErrHandler{handler: func(r *http.Request) (val string, err error) {
		buck_name := mux.Vars(r)["bucket"]
		key := mux.Vars(r)["key"]

		err = bs.db.View(func(tx *bolt.Tx) (err error) {
			if buck := tx.Bucket([]byte(buck_name)); buck != nil {
				val = string(buck.Get([]byte(key)))
				return
			}
			return ErrNoBucket
		})

		return
	}}
}

func DeleteHandler(bs *BoltServer) *ErrHandler {
	return &ErrHandler{handler: func(r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]
		key := mux.Vars(r)["key"]

		return bs.db.Update(func(tx *bolt.Tx) error {
			if buck := tx.Bucket([]byte(buck_name)); buck != nil {
				return buck.Delete([]byte(key))
			}
			return ErrNoBucket
		})
	}}
}
