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
		f, h, err := r.FormFile("file")
		if err != nil {
			log.Println(err)
			return
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println(err)
			return
		}
		outd, err := ioutil.TempDir(datadir, "")
		if err != nil {
			log.Println(err)
			return
		}
		fn := outd+"/"+h.Filename
		err = ioutil.WriteFile(fn, data, 0777)
		if err != nil {
			fmt.Println(err)
		}
		w.Write([]byte("Here is your link: <a href=\"/"+fn+"\">"+h.Filename+"</a>"))
	}
}

func dhandler(w http.ResponseWriter, r *http.Request) {
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/about/", handler)
	http.HandleFunc("/u/", uhandler)
	http.Handle("/c/", captcha.Server(captcha.StdWidth, captcha.StdHeight))


	http.Handle("/data/",
		http.StripPrefix("/data/",
			http.FileServer(http.Dir(datadir))))

	log.Print("Launching on http://localhost:"+*port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

const indexsrc = `
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html;charset=utf-8" >
		<link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css">
		<style type="text/css">
			.btn-file {
				position: relative;
				overflow: hidden;
			}
			.btn-file input[type=file] {
				position: absolute;
				top: 0;
				right: 0;
				min-width: 100%;
				min-height: 100%;
				font-size: 999px;
				text-align: right;
				filter: alpha(opacity=0);
				opacity: 0;
				outline: none;
				background: white;
				cursor: inherit;
				display: block;
			}
			footer {
				font-size:	small;
			}
		</style>
		<title>Temporary file uploader</title>
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
	</head>

	<body>
		<nav class="navbar navbar-default" role="navigation">
			<a class="navbar-brand" href="#">Temporary file uploader</a>
		</nav>
		<div class="container">
			<div class="text-center">
				<form enctype="multipart/form-data" action="/u/" method=post>
					<span class="btn btn-default btn-file">
						Max file size: 5Mo; available 24h<input type="file" name="file" />
					</span>
					<p><img id="image" src="/c/{{.CaptchaId}}.png" alt="Captcha image"></p>
					<p> (Reload for new captcha) </p>
					<input type="hidden" name="captchaId" value="{{.CaptchaId}}" /><br>
					<input type="text" name="captchasol" /><br />
					<input class="btn btn-success btn-lg" type="submit" value="Upload" />
				</form>
			</div>
		</div>

		<footer class="">
			<div class="container text-center">
				<p>Free of use. Consider <a href="TODO">donations.</a> if you enjoy! </p>
				<p>Source available on <a href="https://github.com/heaumer/fup">Github</a></p>
			</div>
		</footer>
	</body>
</html>
`
