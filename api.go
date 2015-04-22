package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

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

type ErrHandlerFunc func(w http.ResponseWriter, r *http.Request) error
type DataErrHandlerFunc func(w http.ResponseWriter, r *http.Request) (string, error)

func jsonResp(hf interface{}) http.HandlerFunc {
	switch hf := hf.(type) {
	case ErrHandlerFunc:
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)

			err := hf(w, r)

			if err != nil {
				encoder.Encode(Response{err.Error()})
				return
			}
			encoder.Encode(Response{})
		}
	case DataErrHandlerFunc:
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)

			data, err := hf(w, r)

			if err != nil {
				encoder.Encode(DataResponse{Response{err.Error()}, ""})
				return
			}
			encoder.Encode(DataResponse{Data: data})
		}
	}
	panic("Illegal type for handler function")
}

//

var (
	ErrNoBucket = errors.New("bucket doesn't exist")
)

func CreateBucketHandler(bs *BoltServer) ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucket([]byte(buck_name))
			return
		})
	}
}

func CreateBucketIfNotExistsHandler(bs *BoltServer) ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucketIfNotExists([]byte(buck_name))
			return
		})
	}
}

func DeleteBucketHandler(bs *BoltServer) ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) error {
			return tx.DeleteBucket([]byte(buck_name))
		})
	}
}

//

func PutHandler(bs *BoltServer) ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
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
	}
}

func GetHandler(bs *BoltServer) DataErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (val string, err error) {
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
	}
}
