package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
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
		r.ParseForm()
		name := r.PostFormValue("name")
		text := r.PostFormValue("text")
		fmt.Println(text)
		// コンテナ内にコピーするための一時ファイルを作成
		filename := "tmp/" + name + ".text"
		filew, err := os.Create(filename)
		if err != nil {
			fmt.Fprint(w, err)
		}
		filew.Write([]byte(text))
		filew.Close()
		// 作成した一時ファイルをコピーする
		err = exec.Command("docker", "cp", filename, name+":/data/text").Run()
		if err != nil {
			http.Redirect(w, r, "/?target=Command&error=cp", 302)
			return
		}
		http.Redirect(w, r, "/", 302)
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
		psaout, err := exec.Command("docker", "ps", "-a", "--format", "{{.Names}},{{.Image}},{{.Ports}}").Output()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		csvReader := csv.NewReader(bytes.NewReader(psaout))
		psaout, err = exec.Command("docker", "ps", "--format", "{{.Names}},{{.Image}},{{.Ports}}").Output()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		// コンテナを表示する
		for {
			v, err := csvReader.Read()
			if err != nil {
				break
			}
			image := v[1]
			if image != "webdocker" {
				continue
			}
			name := v[0]
			// cport ->より左側を取る アクセスできる形式に変換
			cport := v[2]
			running := cport != ""
			text := make([]byte, 0)
			statusClass := "stopped"
			// コンテナが起動していたら
			if running {
				cport = cport[:strings.Index(cport, "->")]
				text, err = exec.Command("docker", "exec", name, "cat", "/data/text").Output()
				if err != nil {
					fmt.Fprint(w, err)
					return
				}
				statusClass = "running"
			} else {
				filer, err := os.Open("tmp/" + name + ".text")
				if err != nil {
					goto skip
				}
				b := new(bytes.Buffer)
				io.Copy(b, filer)
				filer.Close()
				text = b.Bytes()
			}
		skip:

			hiddenName := `<input type="hidden" name="name" value="` + name + `">`
			html += `
<div class="item ` + statusClass + `">
	<h3>` + name + `</h3>
	<hr>`
			// 起動状態によって表示を変える
			if running {
				// 起動している
				html += `
	<form action="/stop" method="post">
		<input type="submit" value="停止" class="stop">
		` + hiddenName + `
	</form>
	<a class="port" href="http://` + cport + `" target="_new">` + cport + `</a>
	<form action="/update" method="post">
		<textarea name="text" cols="30" rows="10">` + string(text) + `</textarea><br>
		` + hiddenName + `
		<input type="submit" value="更新" class="update">
	</form>`
			} else {
				// 停止している
				html += `
	<form action="/start" method="post">
		<input type="submit" value="起動" class="start">
		` + hiddenName + `
	</form>
	<textarea name="text" cols="30" rows="10" readonly>` + string(text) + `</textarea><br>
	<form action="/remove" method="post">
		` + hiddenName + `
		<input type="submit" value="削除" class="remove">
	</form>`
			}

			html += `</div>`
		}
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
