package main

import (
    "log"
    "os"
    "encoding/json"
    "net/http"

    "gopkg.in/mgo.v2"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
)

var config struct {
    WebAddr string `json:"web_addr"`
    DBAddr string `json:"db_addr"`
    DBName string `json:"db_name"`
    DBUser string `json:"db_user"`
    DBPass string `json:"db_pass"`
}

var database *mgo.Database

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
}

func main() {
    initMongo();

    m := martini.Classic()

    m.Use(martini.Static("assets"))
    m.Use(AuthHandler)

    m.Use(render.Renderer(render.Options{
        Layout: "layout",
    }))

    m.Get("/", func(r render.Render, req *http.Request, u User) {
       r.HTML(200, "home", u.LoggedIn())
    })

    m.RunOnAddr(config.WebAddr)
}
