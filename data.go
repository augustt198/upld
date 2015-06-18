package main

import (
    "time"
    "errors"
    "mime/multipart"

    "gopkg.in/mgo.v2/bson"
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

func Upload(file multipart.File, header *multipart.FileHeader,
    u User) (*bson.ObjectId, error) {

    if !u.LoggedIn() {
        return nil, errors.New("You must be logged in to do that")
    }

    dup, err := isDup(header.Filename, u)
    if err != nil {
        return nil, err
    } else if dup {
        return nil, errors.New("Duplicate filename")
    }

    docId := bson.NewObjectId()
    doc := bson.M{
        "_id": docId,
        "user_id": u.OID(),
        "name": header.Filename,
        "created_on": time.Now(),
        "favorite": false,
    }

    path := u.Username() + "/" + header.Filename
    typeHeaders := header.Header["Content-Type"]
    contType := ""
    if  typeHeaders != nil && len(typeHeaders) > 0 {
        contType = typeHeaders[0]
    }

    multi, err := bucket.InitMulti(path, contType, "")

    parts, err := multi.PutAll(file, partSize)
    if err != nil {
        return nil, err
    }
    err = multi.Complete(parts)
    if err != nil {
        return nil, err
    }

    err = database.C("uploads").Insert(doc)
    if err != nil {
        return nil, err
    }

    return &docId, nil
}

func RemoveUpload(u User, name string) bool {
    path := u.Username() + "/" + name

    return bucket.Del(path) == nil
}
