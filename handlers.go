package main

import (
    "io/ioutil"
    "strings"
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
    OpenGraph() map[string]interface{}
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

    m.Post("/delete", deleteSubmit)
    m.Post("/favorite", favoriteSubmit)

    m.Get("/view/:user/:id", viewPage)
}

func NotFound(r http.ResponseWriter) {
    r.WriteHeader(404)
    r.Write([]byte("Not found."))
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
        oid := entry["_id"].(bson.ObjectId)
        newMap := make(bson.M, len(entry))
        for k, v := range entry { newMap[k] = v }

        name := url.QueryEscape(entry["name"].(string))
        
        path := u.Username() + "/" + name
        var thumbnail string
        if b, _ := entry["thumbnail"].(bool); b {
            thumbnail = u.Username() + ".thumbs/" + name
        } else {
            thumbnail = path
        }

        newMap["ImageURL"] = config.StorageBaseURL + path
        newMap["ThumbnailURL"] = config.StorageBaseURL + thumbnail
        newMap["ViewURL"] = "/view/" + u.Username() + "/" + oid.Hex()
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

func uploadSubmit(r render.Render, u User, req *http.Request) (int, string) {
    file, header, err := req.FormFile("upload")
    if err != nil {
        return 400, "Missing file upload"
    }

    id, err := Upload(file, header, u)
    if err != nil {
        return 500, "Upload error"
    }

    return 200, u.Username() + "/" + id.Hex()
}

func deleteSubmit(u User, r *http.Request) (int, string) {
    if !u.LoggedIn() {
        return 403, "Not authorized"
    }

    bytes, err := ioutil.ReadAll(r.Body)

    if err != nil {
        return 400, "Invalid request"
    }

    ids := strings.Split(string(bytes), ",")
    for _, id := range ids {
        if !bson.IsObjectIdHex(id) {
            return 400, "Invalid ID"
        }
    }

    removed := make([]string, 0, len(ids))

    // all IDs are valid
    for _, id := range ids {
        query := bson.M{
            "user_id": u.OID(),
            "_id": bson.ObjectIdHex(id),
        }    

        var result bson.M
        err := database.C("uploads").Find(query).One(&result)
        if err != nil {
            continue
        }
        name := result["name"].(string)
        if !RemoveUpload(u, name) {
            continue
        }
        err = database.C("uploads").Remove(query)
        if err != nil {
            continue
        }
        removed = append(removed, id)
    }

    return 200, strings.Join(removed, ",")
}

func favoriteSubmit(u User, r *http.Request) (int, string) {
    if !u.LoggedIn() {
        return 403, "Not authorized"
    }

    action := r.URL.Query().Get("fav") == "1"
    bytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return 400, "Invalid request"
    }

    ids := strings.Split(string(bytes), ",")

    for _, id := range ids {
        if !bson.IsObjectIdHex(id) {
            return 400, "Invalid ID"
        }
    }

    removed := make([]string, 0, len(ids))

    for _, id := range ids {
        query := bson.M{
            "user_id": u.OID(),
            "_id": bson.ObjectIdHex(id),
        }
        
        update := bson.M{
            "$set": bson.M{"favorite": action},
        }
        err = database.C("uploads").Update(query, update)
        if err == nil {
            removed = append(removed, id)
        }
    }

    return 200, strings.Join(removed, ",")
}

func viewPage(r render.Render, params martini.Params,
    t TemplateData, res http.ResponseWriter) {

    username := params["user"]
    if !bson.IsObjectIdHex(params["id"]) {
        NotFound(res)
        return
    }
    uploadId := bson.ObjectIdHex(params["id"])

    query := bson.M{"username": username}
    var result bson.M
    err := database.C("users").Find(query).One(&result)
    if err != nil {
        NotFound(res)
        return
    }

    user_id := result["_id"]
    query = bson.M{"_id": uploadId, "user_id": user_id}
    err = database.C("uploads").Find(query).One(&result)
    if err != nil {
        NotFound(res)
        return
    }

    path := username + "/" + url.QueryEscape(result["name"].(string))
    t.Data()["Name"] = result["name"]
    t.Data()["S3_URL"] = config.StorageBaseURL + path

    t.OpenGraph()["og:title"] = result["name"]
    t.OpenGraph()["og:image"] = config.StorageBaseURL + path

    r.HTML(200, "view", t)
}
