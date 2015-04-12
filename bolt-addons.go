package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

type BoltServer struct {
	port string
	db   *bolt.DB
}

func NewBoltSrv(port string) *BoltServer {
	bs := new(BoltServer)
	bs.port = port
	return bs
}

//

type Response struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func jsonHandler(hf HandlerFunc) http.HandlerFunc {
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

func CreateBucketHandler(bs *BoltServer) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		buck_name := mux.Vars(r)["bucket"]

		return bs.db.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucket([]byte(buck_name))
			return
		})
	}
}

func (self *BoltServer) Start() error {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	self.db = db

	r := mux.NewRouter()
	r.HandleFunc("/CreateBucket/{bucket}", jsonHandler(CreateBucketHandler(self))).Methods("GET")
	http.Handle("/", r)

	return http.ListenAndServe(":"+self.port, nil)
}

func startBoltSrv(port string) error {
	b := NewBoltSrv(port)

	return b.Start()
}

func main() {
	app := cli.NewApp()
	app.Name = ""
	app.Usage = "bolt http server"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "port, p",
			Value: "3344",
		},
	}
	app.Action = func(c *cli.Context) {
		err := startBoltSrv(c.String("port"))
		if err != nil {
			log.Fatal(err)
		}
	}
	app.Run(os.Args)
}
