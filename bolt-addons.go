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

type Error struct {
	Err string `json:"error"`
}

func CreateBucketHandler(bs *BoltServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		buck_name := mux.Vars(r)["bucket"]

		err := bs.db.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucket([]byte(buck_name))
			return
		})

		if err != nil {
			encoder.Encode(Error{err.Error()})
			return
		}
		encoder.Encode(Error{})
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
	r.HandleFunc("/CreateBucket/{bucket}", CreateBucketHandler(self)).Methods("GET")
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
