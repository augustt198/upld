package main

import (
    "log"
    "net/http"

    "github.com/go-martini/martini"
    "github.com/gorilla/sessions"

    "gopkg.in/mgo.v2/bson"
)

var store = sessions.NewCookieStore([]byte("this is supposed to be secret"))

type User interface {
    LoggedIn() bool
    Username() string
}

type user struct {
    doc *bson.M

    loggedIn bool
}

func (u user) LoggedIn() bool {
    return u.loggedIn
}

func (u user) Username() string {
    if u.doc != nil {
        return (*(u.doc))["username"].(string)
    } else {
        return ""
    }
}

func DefaultUser() User {
    return user{nil, false}
}

func AuthHandler(res http.ResponseWriter, req *http.Request, ctx martini.Context) {
    session, _ := store.Get(req, "users")

    var u User
    log.Print(config)

    oid, ok := session.Values["oid"].(string)
    if ok && bson.IsObjectIdHex(oid) {
        query := bson.M{"_id": bson.ObjectIdHex(oid)}
        var result bson.M
        err := database.C("users").Find(query).One(&result)

        if err != nil {
            u = user{&result, true}
        } else {
            u = DefaultUser()
        }
    } else {
        u = DefaultUser()
    }

    ctx.MapTo(u, (*User)(nil))
}
