package main

import (
    "errors"

    "gopkg.in/mgo.v2/bson"

    "github.com/aws/aws-sdk-go/service/s3"
)

const maxSize uint = 20 * 1000000
const partSize int64 = 10 * 1000000

func isDup(filename string, u User) (bool, error) {
    if !u.LoggedIn() {
        return false, nil
    }

    query := bson.M{
        "user_id": u.OID(),
        "name": filename,
    }

    count, err := database.C("uploads").Find(query).Count()
    if err != nil {
        return false, err
    }
    return count > 1, nil
}

func RemoveMulti(u User, ids []bson.ObjectId) ([]string, error) {
    keys := make(map[string]bson.ObjectId, len(ids))
    objects := make([]*s3.ObjectIdentifier, 0, len(ids))
    doc := bson.M{
        "_id": bson.M{"$in": ids},
        "user_id": u.OID(),
    }
    query := database.C("uploads").Find(doc)
    iter := query.Iter()

    var entry bson.M
    for iter.Next(&entry) {
        key := u.Username() + "/" + entry["name"].(string)
        keys[key] = entry["_id"].(bson.ObjectId)

        obj := s3.ObjectIdentifier{Key: &key}
        objects = append(objects, &obj)
    }

    del := s3.Delete{Objects: objects}
    input := s3.DeleteObjectsInput{
        Bucket: &config.BucketName,
        Delete: &del,
    }

    output, err := storage.DeleteObjects(&input)
    if err != nil {
        return nil, errors.New("Storage error")
    }

    _, err = database.C("uploads").RemoveAll(doc)
    if err != nil {
        return nil, errors.New("Database error")
    }

    removed := make([]string, 0, len(output.Deleted))
    for _, d := range output.Deleted {
        oid, ok := keys[*d.Key]
        if ok {
            removed = append(removed, oid.Hex()) 
        }
    }

    return removed, nil
}
