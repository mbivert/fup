package main

import (
	"flag"
	"github.com/dchest/captcha"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
	"time"
)


const (
	// Maximum storage available (2G)
	Maxstorage = 2<<30
	// File maximum size and storage duration
	Maxsize = 5<<20
	Maxtime = 24*3600
//	Maxtime = 20

	// Cleaning every Cleantime
	Cleantime = 2*time.Hour
//	Cleantime = 2*time.Second


)

var port = flag.String("port", "8080", "Listening HTTP port")
var datadir = flag.String("data", "./data/", "Listening HTTP port")

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

type Fileinfo struct {
	ctime	int64	// creation time as UNIX timestamp
	sz		int		// len in bytes
}

var cache map[string]Fileinfo
var cachesz int64

func cleaning() {
	for {
		time.Sleep(Cleantime)
		now := time.Now().Unix()
		for k, _ := range cache {
			if now-cache[k].ctime >= Maxtime {
				err := os.RemoveAll(path.Dir(k))
				if err != nil {
					log.Println(err)
				}
				cachesz -= int64(cache[k].sz)
				delete(cache, k)
			}
		}
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r := io.LimitReader(f, Maxsize+1)
		data, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(data) >= Maxsize {
			log.Println("attempting to store a", len(data), "bytes long file")
			http.Error(w, "File too big", http.StatusRequestEntityTooLarge)
			return
		}

		if cachesz+int64(len(data)) >= Maxstorage {
			http.Error(w, "Maximum storage reach; please wait", http.StatusRequestEntityTooLarge)
			return
		}

		outd, err := ioutil.TempDir(*datadir, "")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fn := path.Clean(outd+"/"+h.Filename)
		err = ioutil.WriteFile(fn, data, 0777)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cache[fn] = Fileinfo{ time.Now().Unix(), len(data) }
		w.Write([]byte("Here is your link: <a href=\"/"+fn+"\">"+h.Filename+"</a>"))
	}
}

func main() {
	flag.Parse()

	cache = make(map[string]Fileinfo)
	cachesz = 0

	go cleaning()

	http.HandleFunc("/", handler)
	http.HandleFunc("/about/", handler)
	http.HandleFunc("/u/", uhandler)
	http.Handle("/c/", captcha.Server(captcha.StdWidth, captcha.StdHeight))


	http.Handle("/data/",
		http.StripPrefix("/data/",
			http.FileServer(http.Dir(*datadir))))

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
				<p>Source available on <a href="https://github.com/heaumer/fup">Github</a></p>
			</div>
		</footer>
	</body>
</html>
`
