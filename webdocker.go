package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/intelfike/checkmodfile"
)

var port = flag.String("http", ":8888", "HTTP port number.")
var nameReg = regexp.MustCompile("^[a-zA-Z0-9]+$")

func init() {
	handleFunc("/create", http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.PostFormValue("name")
		port := r.PostFormValue("port")
		if name == "" || port == "" {
			http.Redirect(w, r, "/?target=Form&error=empty", 302)
			return
		}
		_, err := strconv.Atoi(port)
		if err != nil {
			http.Redirect(w, r, "/?target=Port&error=must use only number.", 302)
			return
		}
		if !nameReg.MatchString(name) {
			http.Redirect(w, r, "/?target=Name&error=must use only character a-z,A-Z,0-9", 302)
			return
		}
		err = exec.Command("docker", "run", "-d", "-p", port+":8888", "--name", name, "webdocker", "/server").Run()
		if err != nil {
			http.Redirect(w, r, "/?target=Command&error=run", 302)
			exec.Command("docker", "rm", "-f", name).Run()
			return
		}

		http.Redirect(w, r, "/", 302)
	})
	handleFunc("/remove", http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.PostFormValue("name")
		err := exec.Command("docker", "rm", "-f", name).Run()
		if err != nil {
			http.Redirect(w, r, "/?target=Command&error=remove", 302)
			return
		}

		http.Redirect(w, r, "/", 302)
	})
	handleFunc("/update", http.MethodPost, func(w http.ResponseWriter, r *http.Request) {

	})
	handleFunc("/start", http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.PostFormValue("name")
		err := exec.Command("docker", "start", name).Run()
		if err != nil {
			http.Redirect(w, r, "/?target=Command&error=start", 302)
			return
		}

		http.Redirect(w, r, "/", 302)
	})
	handleFunc("/stop", http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.PostFormValue("name")
		err := exec.Command("docker", "stop", name).Run()
		if err != nil {
			http.Redirect(w, r, "/?target=Command&error=stop", 302)
			return
		}

		http.Redirect(w, r, "/", 302)
	})
	f, err := checkmodfile.RegistFile("data/index.html")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	handleFunc("/", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		b, err := f.GetBytes()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		html := ""
		out, err := exec.Command("docker", "ps", "-a").Output()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		html += "<pre>" + string(out) + "</pre>"
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		doc.Find("main").SetHtml(html)
		s, err := doc.Html()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		w.Write([]byte(s))
	})
}

func handleFunc(path, method string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			fmt.Fprint(w, "あなたにアクセス権はありません！")
			return
		}
		if r.Method != method {
			fmt.Fprint(w, r.Method, " is bad method")
			return
		}
		handler(w, r)
	})
}

func main() {
	fmt.Println("Start HTTP Server localhost", *port)
	fmt.Println(http.ListenAndServe(*port, nil))
}
