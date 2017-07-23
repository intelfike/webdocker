package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/intelfike/checkmodfile"
)

var port = flag.String("http", ":8888", "HTTP port number.")

func init() {
	f, err := checkmodfile.RegistFile("data/text")
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			err := f.WriteTo(w)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		}
	})
}

func main() {
	fmt.Println("Start HTTP Server localhost", *port)
	fmt.Println(http.ListenAndServe(*port, nil))
}
