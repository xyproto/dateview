package main

// a simple calendar that uses moskus, negroni and permissions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/xyproto/moskus"
	//"github.com/xyproto/permissions"
	"github.com/codegangsta/negroni"
)

func main() {
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

	n := negroni.Classic()

	// Middleware goes here
	//n.Use(moose.NewMiddleware())

	// Handler goes last
	n.UseHandler(mux)

	// Serve
	n.Run(":9030")
}
