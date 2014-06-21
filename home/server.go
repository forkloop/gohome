package main

import (
    "os"
    "github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func main() {
    martini.Env = martini.Prod

    m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Layout: "base",
	}))

	// default public
	m.Use(martini.Static("assets"))

    m.Get("/", func(render render.Render) {
		render.HTML(200, "home", nil)
    })

    os.Setenv("PORT", "80")
    m.Run()
}
