package main

import (
    "log"
    "os"
    "encoding/json"

    "gopkg.in/mgo.v2"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/sqs"
)

var config struct {
    WebAddr string `json:"web_addr"`
    
    DBAddr string `json:"db_addr"`
    DBName string `json:"db_name"`
    DBUser string `json:"db_user"`
    DBPass string `json:"db_pass"`

    BucketName string `json:"bucket_name"`
    ImageBaseURL string `json:"image_base_url"`
    ThumbnailBaseURL string `json:"thumbnail_base_url"`

    ThumbsQueueName string `json:"thumbs_queue_name"`
    ThumbsQueueRegion string `json:"thumbs_queue_region"`
    ThumbsQueueURL string

    AWSConfig *aws.Config
}

var database *mgo.Database
var storage *s3.S3
var queue *sqs.SQS

func initMongo() {
    cfg, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    parser := json.NewDecoder(cfg)
    if err = parser.Decode(&config); err != nil {
        log.Fatal("Could not decode json: ", err)
    }
    config.AWSConfig = aws.DefaultConfig

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

func initStorage() {
    storage = s3.New(&aws.Config{Region: "us-east-1"})

    log.Print("S3 connected")
}

func initThumbsQueue() {
    queue = sqs.New(&aws.Config{
        Region: config.ThumbsQueueRegion,
    })

    input := sqs.GetQueueURLInput{
        QueueName: &config.ThumbsQueueName,
    }
    output, err := queue.GetQueueURL(&input)
    if err != nil {
        log.Fatal(err)
    }
    config.ThumbsQueueURL = *output.QueueURL

    log.Print("Thumbnail queue connected")
}

func main() {
    initMongo()
    initStorage()
    initThumbsQueue()

    m := martini.Classic()

    m.Use(martini.Static("assets"))
    m.Use(AuthHandler)

    m.Use(render.Renderer(render.Options{
        Layout: "layout",
    }))

    RegisterHandlers(m)

    m.RunOnAddr(config.WebAddr)
}
