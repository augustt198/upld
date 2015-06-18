package main

import (
    "log"
    "os"
    "encoding/json"

    "gopkg.in/mgo.v2"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"

    "github.com/mitchellh/goamz/aws"
    "github.com/mitchellh/goamz/s3"
)

var config struct {
    WebAddr string `json:"web_addr"`
    
    DBAddr string `json:"db_addr"`
    DBName string `json:"db_name"`
    DBUser string `json:"db_user"`
    DBPass string `json:"db_pass"`

    BucketName string `json:"bucket_name"`

    StorageBaseURL string
}

var database *mgo.Database
var bucket *s3.Bucket

func initMongo() {
    cfg, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    parser := json.NewDecoder(cfg)
    if err = parser.Decode(&config); err != nil {
        log.Fatal("Could not decode json: ", err)
    }

    session, err := mgo.DialWithInfo(&mgo.DialInfo{
        Addrs: []string{config.DBAddr},
        Database: config.DBName,
        Username: config.DBUser,
        Password: config.DBPass,
    })
    if err != nil {
        log.Fatal(err)
    }
    database = session.DB(config.DBName)
    log.Print("Database connected")
}

func initAws() {
    auth, err := aws.EnvAuth()
    if err != nil {
        log.Fatal(err)
    }

    client := s3.New(auth, aws.USEast)
    bucket = client.Bucket(config.BucketName)

    config.StorageBaseURL = "https://s3.amazonaws.com/" +
        config.BucketName + "/"

    log.Print("S3 connected")
}

func main() {
    initMongo()
    initAws()

    m := martini.Classic()

    m.Use(martini.Static("assets"))
    m.Use(AuthHandler)

    m.Use(render.Renderer(render.Options{
        Layout: "layout",
    }))

    RegisterHandlers(m)

    m.RunOnAddr(config.WebAddr)
}
