package main

import (
	"net/http"

	. "github.com/codegangsta/negroni"
	"github.com/unrolled/render"
)

func main() {
	r := render.New(render.Options{})

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		data := map[string]string{
			"version_string": version_string,
			"server":         default_server,
			"status":         "up",
			"shortname":      "dx",
		}
		if DefaultServerUp() {
			data["status"] = "up"
			data["statuscolor"] = "#a0a0ff"
		} else {
			data["status"] = "down"
			data["statuscolor"] = "#ffa0a0"
		}

		// Reload template
		r = render.New(render.Options{})

		// Render and return
		r.HTML(w, http.StatusOK, "index", data)
	})

	n := Classic()

	// Share the files in static
	n.Use(NewStatic(http.Dir("static")))

	// Handler goes last
	n.UseHandler(mux)

	// Serve
	n.Run(":80")
}
