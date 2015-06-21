package main

import (
    "log"
    "os"
    "time"
    "errors"
    "strings"
    "strconv"
    "encoding/json"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/sqs"

    "bytes"
    "image"
    "image/png"
    "github.com/nfnt/resize"

    _ "image/jpeg"
    _ "image/gif"
)

var config struct {
    DBAddr string `json:"db_addr"`
    DBName string `json:"db_name"`
    DBUser string `json:"db_user"`
    DBPass string `json:"db_pass"`

    BucketName string `json:"bucket_name"`

    ThumbsQueueName string `json:"thumbs_queue_name"`
    ThumbsQueueRegion string `json:"thumbs_queue_region"`
    ThumbsQueueURL string
}

var database *mgo.Database
var storage *s3.S3
var queue *sqs.SQS

func initMongo() {
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
    storage = s3.New(&aws.Config{
        Region: "us-east-1",
    })

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
    cfg, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    parser := json.NewDecoder(cfg)
    if err = parser.Decode(&config); err != nil {
        log.Fatal(err)
    }

    initMongo()
    initStorage()
    initThumbsQueue()

    loop()
}

func loop() {
    input := sqs.ReceiveMessageInput{
        MaxNumberOfMessages: aws.Long(10),
        QueueURL: &config.ThumbsQueueURL,
        VisibilityTimeout: aws.Long(2 * 60),
    }
    for {
        output, err := queue.ReceiveMessage(&input)
        if err != nil {
            log.Print(err)
        } else {
            for _, msg := range output.Messages {
                go process(msg)
            }            
        }
        time.Sleep(3 * time.Second)
    }
}

type ThumbnailRequest struct {
    UploadID bson.ObjectId
    MaxWidth int
    MaxHeight int
}

func parseMessage(msg *sqs.Message) (*ThumbnailRequest, error) {
    parts := strings.Split(*msg.Body, ",")

    if len(parts) != 3 {
        return nil, errors.New("Expected 3 parts")
    }
    if !bson.IsObjectIdHex(parts[0]) {
        return nil, errors.New("Expected first part to an ObjectId")
    }
    id := bson.ObjectIdHex(parts[0])

    maxWidth, err := strconv.Atoi(parts[1])
    if err != nil {
        return nil, errors.New("Expected second part to be an int")
    }
    maxHeight, err := strconv.Atoi(parts[2])
    if err != nil {
        return nil, errors.New("Expected third part to be an int")
    }

    req := ThumbnailRequest{
        UploadID: id,
        MaxWidth: maxWidth,
        MaxHeight: maxHeight,
    }

    return &req, nil
}

func getUsername(id interface{}) (string, bool) {
    var user bson.M

    if database.C("users").FindId(id).One(&user) != nil {
        return "", false
    } else {
        return user["username"].(string), true
    }
}

func process(msg *sqs.Message) {
    del := sqs.DeleteMessageInput{
        QueueURL: &config.ThumbsQueueURL,
        ReceiptHandle: msg.ReceiptHandle,
    }
    _, err := queue.DeleteMessage(&del)
    if err != nil {
        log.Print("COULD NOT DELETE MESSAGE: ", err)
    }

    req, err := parseMessage(msg)
    if err != nil {
        log.Print(err)
        return
    }

    var result bson.M
    query := database.C("uploads").FindId(req.UploadID)
    if query.One(&result) != nil {
        log.Print("No upload with id ", req.UploadID, " found")
        return
    }

    username, ok := getUsername(result["user_id"])
    if !ok {
        log.Print("Username not found")
        return
    }

    key := username + "/" + result["name"].(string)
    input := s3.GetObjectInput{
        Bucket: &config.BucketName,
        Key: &key,
    }
    output, err := storage.GetObject(&input)
    if err != nil {
        log.Print(err)
        return
    }

    img, _, err := image.Decode(output.Body)
    if err != nil {
        log.Print(err)
        return
    }

    thumb := resize.Thumbnail(
        uint(req.MaxWidth), uint(req.MaxHeight),
        img, resize.NearestNeighbor)

    var buf bytes.Buffer
    err = png.Encode(&buf, thumb)
    if err != nil {
        log.Print(err)
        return
    }
    r := bytes.NewReader(buf.Bytes())

    thumbKey := username + ".thumbs/" + result["name"].(string)
    put := s3.PutObjectInput{
        Bucket: &config.BucketName,
        Key: &thumbKey,
        Body: r,
    }
    _, err = storage.PutObject(&put)

    if err != nil {
        log.Print(err)
        return
    } else {
        log.Print("Saved thumbnail")
    }

    update := bson.M{
        "$set": bson.M{"thumbnail": true},
    }

    err = database.C("uploads").UpdateId(req.UploadID, update)
    if err != nil {
        log.Print(err)
    }

}
