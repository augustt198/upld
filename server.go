package main

import (
  "github.com/go-martini/martini"
  "github.com/martini-contrib/render"
)

func main() {
  m := martini.Classic()

  m.Use(render.Renderer(render.Options{
    Layout: "layout",
  }))

  m.Get("/", func(r render.Render) {
    r.HTML(200, "layout", nil)
  })

  m.Run()
}
