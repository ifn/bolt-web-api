package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

type Response struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type ErrHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func jsonResp(hf ErrHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		err := hf(w, r)

		if err != nil {
			encoder.Encode(Response{1, err.Error()})
			return
		}
		encoder.Encode(Response{})
	}
}

//

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

type Message struct {
	Key   string `json:key`
	Value string `json:value`
}

func PutHandler(bs *BoltServer) ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		var m Message
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			return err
		}

		return bs.db.Update(func(tx *bolt.Tx) error {
			if buck := tx.Bucket([]byte(buck_name)); buck != nil {
				return buck.Put([]byte(m.Key), []byte(m.Value))
			}
			return errors.New("bucket doesn't exist")
		})
	}
}
