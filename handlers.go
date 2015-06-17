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
    
    m.Get("/login", RequireNoAuth, loginPage)
    m.Post("/login", RequireNoAuth, loginSubmit)

    m.Post("/logout", RequireAuth, logoutSubmit)

    m.Get("/register", RequireNoAuth, registerPage)
    m.Post("/register", RequireNoAuth, registerSubmit)

    m.Get("/me", RequireAuth, mePage)

    m.Get("/upload", RequireAuth, uploadPage)
    m.Post("/upload", RequireAuth, uploadSubmit)
}

func RequireAuth(u User, ren render.Render, r *http.Request,
    w http.ResponseWriter) {

    if !u.LoggedIn() {
        AddFlash(r, w, "You must be logged in to view that page")
        ren.Redirect("/login")
    }
}

func RequireNoAuth(u User, ren render.Render, r *http.Request,
    w http.ResponseWriter) {

    if u.LoggedIn() {
        AddFlash(r, w, "Already logged in")
        ren.Redirect("/")
    }
}

func homePage(r render.Render, u User, t TemplateData) {
    r.HTML(200, "home", t)
}


func loginPage(r render.Render, req *http.Request,
    w http.ResponseWriter, u User, t TemplateData) {


    r.HTML(200, "login", t)
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

    UserLogout(req, w)
    AddFlash(req, w, "Logged out")
    r.Redirect("/")
}

func registerPage(r render.Render, req *http.Request,
    w http.ResponseWriter, u User, t TemplateData) {

    r.HTML(200, "register", t)
}

func registerSubmit(r render.Render, u User, req *http.Request,
    w http.ResponseWriter) {

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

func mePage(r render.Render, u User, req *http.Request,
    w http.ResponseWriter, t TemplateData) {

    r.HTML(200, "me", t)
}

func uploadPage(r render.Render, u User, req *http.Request,
    w http.ResponseWriter, t TemplateData) {

   r.HTML(200, "upload", t)
}

func uploadSubmit(r render.Render, u User, req *http.Request) string {
    file, header, err := req.FormFile("upload")
    if err != nil {
        return err.Error()
    }

    err = Upload(file, header, u)
    if err != nil {
        return err.Error()
    }

    return "success"
}
