package main

import (
    "math"
    "strconv"
    "net/url"
    "net/http"
    "html/template"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"

    "gopkg.in/mgo.v2/bson"
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

    var page int
    page, err := strconv.Atoi(req.URL.Query().Get("page"))
    if err != nil {
        page = 1
    }

    doc := bson.M{"user_id": u.OID()}
    query := database.C("uploads").Find(doc).Sort("-created_on")
    total, err := query.Count()
    var pages int
    if err != nil {
        pages = 1
    } else {
        pages = int(math.Ceil(float64(total) / 20))
    }
    list := make([]bson.M, 0, 20)
    iter := Paginate(page, 20, query).Iter()
    var entry bson.M
    for i := 0; iter.Next(&entry); i++ {
        newMap := make(bson.M, len(entry))
        for k, v := range entry { newMap[k] = v }

        path := u.Username() + "/" + url.QueryEscape(entry["name"].(string))
        newMap["S3_URL"] = config.StorageBaseURL + path
        list = append(list, newMap)
    }
    t.Data()["Uploads"] = list
    t.Data()["TotalUploads"] = total

    t.Data()["Paginate"] = template.HTML(PaginateBar(page, pages))

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
