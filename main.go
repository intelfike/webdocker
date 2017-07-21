package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/intelfike/checkmodfile"
)

var port = flag.String("http", ":8888", "HTTP port number.")

func init() {
	f, err := checkmodfile.RegistFile("data/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			fmt.Fprint(w, "あなたにアクセス権はありません！")
			return
		}
		switch r.Method {
		case http.MethodGet:
			err := f.WriteTo(w)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		case http.MethodPost:
			r.ParseForm()
			name := r.PostFormValue("name")
			port := r.PostFormValue("port")
			if name == "" || port == "" {
				http.Redirect(w, r, "/?error=empty", 302)
				return
			}
			_, err := strconv.Atoi(port)
			if err != nil {
				http.Redirect(w, r, "/?error=Port must be number.", 302)
				return
			}
			cmd, err := exec.Command("docker", "run", "-p=80:"+port, "--name="+name, "webdocker", "init")
			http.Redirect(w, r, "/", 302)
		}
	})
}

func main() {
	fmt.Println("Start HTTP Server localhost", *port)
	fmt.Println(http.ListenAndServe(*port, nil))
}
