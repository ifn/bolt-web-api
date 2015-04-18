package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

type Conf struct {
	Port int

	FilePath string
	FileMode string
}

type BoltServer struct {
	port string

	filePath string
	fileMode os.FileMode

	db *bolt.DB
}

func NewBoltSrv(conf Conf) (bs *BoltServer, err error) {
	bs = new(BoltServer)

	bs.port = strconv.Itoa(conf.Port)
	bs.filePath = conf.FilePath
	fileMode, err := strconv.ParseUint(conf.FileMode, 8, 32)
	if err != nil {
		return
	}
	bs.fileMode = os.FileMode(fileMode)

	return
}

func (self *BoltServer) Start() error {
	db, err := bolt.Open(self.filePath, self.fileMode, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	self.db = db

	r := mux.NewRouter()
	r.HandleFunc("/CreateBucket/{bucket}", jsonResp(CreateBucketHandler(self))).Methods("GET")
	r.HandleFunc("/CreateBucketIfNotExists/{bucket}", jsonResp(CreateBucketIfNotExistsHandler(self))).Methods("GET")
	r.HandleFunc("/DeleteBucket/{bucket}", jsonResp(DeleteBucketHandler(self))).Methods("GET")
	http.Handle("/", r)

	return http.ListenAndServe(":"+self.port, nil)
}

func startBoltSrv(confPath string) (err error) {
	var conf Conf
	if _, err = toml.DecodeFile(confPath, &conf); err != nil {
		return
	}

	b, err := NewBoltSrv(conf)
	if err != nil {
		return
	}

	return b.Start()
}

func main() {
	app := cli.NewApp()
	app.Name = "bwa"
	app.Usage = "bolt web api"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, conf, cfg, c",
			Value: "conf.toml",
		},
	}
	app.Action = func(c *cli.Context) {
		err := startBoltSrv(c.String("config"))
		if err != nil {
			log.Fatal(err)
		}
	}
	app.Run(os.Args)
}
