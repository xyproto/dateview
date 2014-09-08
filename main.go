package main

// a simple calendar that uses moskus, negroni and permissions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/unrolled/render"
	"github.com/xyproto/moskus"
	"github.com/xyproto/permissions"
)

const (
	version_string = "Calendar 0.1"
)

func main() {
	r := render.New(render.Options{})

	perm := permissions.New()
	userstate := perm.UserState()

	//cal, err := moskus.NewCalendar("nb_NO", true)
	//if err != nil {
	//	panic("Could not create a Norwegian calendar!")
	//}

	year := time.Now().Year()

	mux := http.NewServeMux()

	mux.HandleFunc("/easter", func(w http.ResponseWriter, req *http.Request) {
		// When is easter this year?
		easter := moskus.EasterDay(year)
		fmt.Fprintf(w, "Easter %d is at %s\n", year, easter.String()[:10])
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		easter := moskus.EasterDay(year)
		username := userstate.GetUsername(req) // fetch the username as stored in the browser cookie, or blank
		username_desc := "Username:"
		if username == "" {
			username_desc = "No username"
		}
		data := map[string]string{
			"title": version_string,          // application title and version
			"year":           fmt.Sprintf("%d", year), // which year is it?
			"easter_date":    easter.String()[:10],    // when is easter this year?
			"username":       username,
			"username_desc":  username_desc,
		}

		// TODO: Move these two out of the HandleFunc once it the development is done

		// Reload template
		r = render.New(render.Options{})

		// Render and return
		r.HTML(w, http.StatusOK, "index", data)
	})

	n := negroni.Classic()

	perm.SetDenyFunction(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<b>Permission denied!</b>")
	})

	// Middleware goes here

	// Not every user has permissions for everything
	n.Use(perm)

	// Share the files in static
	n.Use(negroni.NewStatic(http.Dir("static")))

	// Handler goes last
	n.UseHandler(mux)

	// Serve
	n.Run(":9030")
}
