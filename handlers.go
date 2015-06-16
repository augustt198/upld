package main

import (
    "net/http"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
)

type TemplateData interface {
    User() User
    Flashes() []interface{}
    Data() map[string]interface{}
}

func RegisterHandlers(m *martini.ClassicMartini) {
    m.Get("/", homePage)
    
    m.Get("/login", loginPage)
    m.Post("/login", loginSubmit)

    m.Post("/logout", logoutSubmit)

    m.Get("/register", registerPage)
    m.Post("/register", registerSubmit)

    m.Get("/me", mePage)
}

func homePage(r render.Render, u User, t TemplateData) {
    r.HTML(200, "home", t)
}


func loginPage(r render.Render, req *http.Request,
    w http.ResponseWriter, u User, t TemplateData) {

    if u.LoggedIn() {
        AddFlash(req, w, "Already logged in")
        r.Redirect("/")
    } else {
        r.HTML(200, "login", t)        
    }
}

func loginSubmit(r render.Render, u User, req *http.Request,
    w http.ResponseWriter) {

    username := req.PostFormValue("username")
    password := req.PostFormValue("password")
    if username != "" && UserAuth(username, password, req, w) {
        AddFlash(req, w, "Successfully logged in")
        r.Redirect("/")
    } else {
        AddFlash(req, w, "Incorrect username or password")
        r.Redirect("/login")
    }
}

func logoutSubmit(r render.Render, u User, req *http.Request,
    w http.ResponseWriter) {

    if !u.LoggedIn() {
        AddFlash(req, w, "Not logged in")
    } else {
        UserLogout(req, w)
        AddFlash(req, w, "Logged out")
    }
    r.Redirect("/")
}

func registerPage(r render.Render, req *http.Request,
    w http.ResponseWriter, u User, t TemplateData) {

    if u.LoggedIn() {
        AddFlash(req, w, "Already logged in")
        r.Redirect("/")
    } else {
        r.HTML(200, "register", t)
    }
}

func registerSubmit(r render.Render, u User, req *http.Request,
    w http.ResponseWriter) {

    if u.LoggedIn() {
        AddFlash(req, w, "Already logged in")
        r.Redirect("/")
    } else {
        usr := req.PostFormValue("username")
        pwd := req.PostFormValue("password")
        pwdConfirm := req.PostFormValue("password_confirmation")

        oid, err := UserRegister(usr, pwd, pwdConfirm)

        if err != nil {
            AddFlash(req, w, err.Error())
            r.Redirect("/register")
            return
        }

        session, _ := store.Get(req, "users")
        session.Values["oid"] = oid.Hex()
        session.Save(req, w)

        AddFlash(req, w, "Successfully registered")
        r.Redirect("/")
    }
}

func mePage(r render.Render, u User, req *http.Request,
    w http.ResponseWriter, t TemplateData) {

    if !u.LoggedIn() {
        AddFlash(req, w, "You must be logged in to see that page")
        r.Redirect("/login")
    } else {
        r.HTML(200, "me", t)
    }
}
