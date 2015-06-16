package main

import (
    "errors"
    "strings"
    "net/http"

    "golang.org/x/crypto/bcrypt"

    "github.com/go-martini/martini"
    "github.com/gorilla/sessions"

    "gopkg.in/mgo.v2/bson"
)

var store = sessions.NewCookieStore([]byte("this is supposed to be secret"))

func AddFlash(req *http.Request, w http.ResponseWriter, val interface{}) {
    session, _ := store.Get(req, "users")
    session.AddFlash(val)

    session.Save(req, w)
}

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


type templatedata struct {
    user User
    flashes []interface{}
    data map[string]interface{}
}

func (t templatedata) User() User {
    return t.user
}

func (t templatedata) Flashes() []interface{} {
    return t.flashes
}

func (t templatedata) Data() map[string]interface{} {
    return t.data
}


func AuthHandler(res http.ResponseWriter, req *http.Request, ctx martini.Context) {
    session, _ := store.Get(req, "users")

    var u User

    oid, ok := session.Values["oid"].(string)

    if ok && bson.IsObjectIdHex(oid) {
        var result bson.M
        id := bson.ObjectIdHex(oid)
        err := database.C("users").FindId(id).One(&result)

        if err != nil {
            u = DefaultUser()
        } else {
            u = user{&result, true}
        }
    } else {
        u = DefaultUser()
    }

    t := templatedata {
        u,
        session.Flashes(),
        make(map[string]interface{}),
    }

    session.Save(req, res)

    ctx.MapTo(u, (*User)(nil))
    ctx.MapTo(t, (*TemplateData)(nil))
}

func UserAuth(username string, password string,
    req *http.Request, res http.ResponseWriter) bool {
    query := bson.M{"username": username}

    var result bson.M
    err := database.C("users").Find(query).One(&result)
    if err != nil {
        return false
    }

    bytes := []byte(password)
    hashed := []byte(result["password"].(string))
    err = bcrypt.CompareHashAndPassword(hashed, bytes)
    
    if err == nil {
        session, _ := store.Get(req, "users")
        session.Values["oid"] = result["_id"].(bson.ObjectId).Hex()
        session.Save(req, res)
        return true
    }

    return false
}

func UserLogout(req *http.Request, res http.ResponseWriter) {
    session, _ := store.Get(req, "users")
    session.Values["oid"] = nil
    session.Save(req, res)
}

func UserRegister(usr string, pwd string, pwdConfirm string) (*bson.ObjectId, error) {

    list := make([]string, 0, 3)
    if usr == "" {
        list = append(list, "username")
    }
    if pwd == "" {
        list = append(list, "password")
    }
    if pwdConfirm == "" {
        list = append(list, "password confirmation")
    }

    if len(list) > 0 {
        msg := "Missing fields: " + strings.Join(list, ", ")
        return nil, errors.New(msg)
    }

    if pwd != pwdConfirm {
        return nil, errors.New("Passwords do not match")
    }

    query := bson.M{"username": usr}
    var result bson.M
    err := database.C("users").Find(query).One(&result)
    
    // found document with username
    if err == nil {
        return nil, errors.New("Username taken")
    }

    hashed, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
    oid := bson.NewObjectId()
    doc := bson.M{
        "_id": oid,
        "username": usr,
        "password": string(hashed),
    }

    err = database.C("users").Insert(doc)
    if err != nil {
        return nil, errors.New("Database error")
    }

    return &oid, nil
}
