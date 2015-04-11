package main

import (
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

type BoltServer struct {
	port string
}

func NewBoltSrv(port string) *BoltServer {
	return &BoltServer{port}
}

func boltHandler(b *BoltServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (self *BoltServer) Start() error {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", boltHandler(self))
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
