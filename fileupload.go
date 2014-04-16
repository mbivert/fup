package main

import (
	"flag"
	"fmt"
	"github.com/dchest/captcha"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

const datadir = "./data/"

var port = flag.String("port", "8080", "Listening HTTP port")

var indextmpl = template.Must(template.New("example").Parse(indexsrc))

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("no."))
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return;
	}

	d := struct { CaptchaId string }{ captcha.New() }

	if err := indextmpl.Execute(w, &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("no."))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchasol")) {
		w.Write([]byte("<p>Bad captcha; try again. </p>"))
	} else {
		file, h, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile(datadir+"/"+h.Filename, data, 0777)
		if err != nil {
			fmt.Println(err)
		} 
	}
}

func dhandler(w http.ResponseWriter, r *http.Request) {
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/u/", uhandler)
	http.Handle("/c/", captcha.Server(captcha.StdWidth, captcha.StdHeight))

	http.HandleFunc("/d/", dhandler)

	log.Print("Launching on http://localhost:"+*port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

const indexsrc = `
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html;charset=utf-8" >
		<title>Temporary file uploader</title>
	</head>

	<body>
		<script>
			function setSrcQuery(e, q) {
				var src = e.src;
				var p = src.indexOf('?');
				if (p >= 0) {
					src = src.substr(0, p);
				}
				e.src = src + "?" + q
			}
	
			function reload() {
				setSrcQuery(document.getElementById('image'), "reload=" + (new Date()).getTime());
				return false;
			}
		</script>
		<form enctype="multipart/form-data" action="/u/" method=post>
			<input type="file" name="file" /> 
			<p> Type the numbers you see in the picture below. </p>
			<p><img id="image" src="/c/{{.CaptchaId}}.png" alt="Captcha image"></p>
			<input type="hidden" name="captchaId" value="{{.CaptchaId}}" /><br>
			<input type="text" name="captchasol" />
			<input type="submit" value="Upload" />
		</form>
	</body>
</html>
`
