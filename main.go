package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

var fndir = "."

func handleDoc(fn string) ([]byte, error) {
	docpath := path.Join(fndir, fn, "doc")
	out, err := os.ReadFile(docpath)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

func handleExec(fn string, body io.Reader) io.Reader {
	fnpath := path.Join(fndir, fn, "handle")

	logpath := path.Join(fndir, fn, "log")
	logfile, err := os.OpenFile(logpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	cmd := exec.Cmd{
		Path:   fnpath,
		Env:    []string{},
		Stdin:  body,
		Stderr: logfile,
	}

	out, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	return out

}

func handle(w http.ResponseWriter, req *http.Request) {
	fn := req.URL.Path
	log.Printf("%v\t%v\n", req.Method, fn)

	if _, err := os.Stat(path.Join(fndir, fn)); os.IsNotExist(err) {
		http.NotFound(w, req)
		return
	}

	switch req.Method {

	case http.MethodGet:
		out, err := handleDoc(fn)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		w.Write(out)

	case http.MethodPost:
		body := req.Body
		start := time.Now()
		out, err := io.ReadAll(handleExec(fn, body))
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		elapsed := time.Since(start).Milliseconds()
		w.Header().Set("X-Duration-Milliseconds", strconv.FormatInt(elapsed, 10))
		w.Write(out)

	default:
		http.Error(w, "", http.StatusBadRequest)
		return

	}
}

func handlePipe(w http.ResponseWriter, req *http.Request) {
	var in io.Reader = req.Body
	var out io.Reader

	fns := strings.Split(req.URL.Path, "/")[2:]
	log.Printf("%v\t/pipe[%s]\n", req.Method, strings.Join(fns, "|"))

	start := time.Now()
	for _, fn := range fns {
		out = handleExec(fn, in)
		in = io.NopCloser(out)
	}
	elapsed := time.Since(start).Milliseconds()
	w.Header().Set("X-Duration-Milliseconds", strconv.FormatInt(elapsed, 10))

	res, err := io.ReadAll(out)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func main() {
	if val, ok := os.LookupEnv("HTTPIPE_DIR"); ok {
		fndir = val
	}
	flag.StringVar(&fndir, "d", fndir, "function directory")
	flag.Parse()

	http.HandleFunc("/", handle)
	http.HandleFunc("/pipe/", handlePipe)
	http.ListenAndServe(":8080", nil)
}
