package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var fndir = "."

func handleDoc(fn string) string {
	docpath := path.Join(fndir, fn, "doc")
	out, err := os.ReadFile(docpath)

	if err != nil {
		panic(err)
	}

	return string(out)

}

func handleExec(fn string, body io.Reader) string {
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

	out, err := cmd.Output()

	if err != nil {
		panic(err)
	}

	return string(out)

}

func handle(w http.ResponseWriter, req *http.Request) {
	fn := req.URL.Path
	fmt.Printf("%v\t%v\n", req.Method, fn)

	if _, err := os.Stat(path.Join(fndir, fn)); os.IsNotExist(err) {
		http.NotFound(w, req)
		return
	}

	switch req.Method {
	case http.MethodGet:
		fmt.Fprintf(w, handleDoc(fn))
	case http.MethodPost:
		body := req.Body
		start := time.Now()
		out := handleExec(fn, body)
		elapsed := time.Since(start).Milliseconds()
		w.Header().Set("X-Duration-Milliseconds", fmt.Sprintf("%d", elapsed))
		fmt.Fprintf(w, out)
	default:
		fmt.Println("unhandled, error")
	}
}

func handlePipe(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("%v\tpipe\n", req.Method)
	fns := strings.Split(req.URL.Path, "/")

	in := req.Body
	out := ""

	start := time.Now()
	// for i := 2; i < len(fns); i++ {
	// 	fn := fns[i]
	for _, fn := range fns {
		if fn == "" || fn == "pipe" {
			continue
		}
		out = handleExec(fn, in)
		in = io.NopCloser(strings.NewReader(out))
		fmt.Printf("\tpipe\t%s\t%s\n", fn, out)
	}
	elapsed := time.Since(start).Milliseconds()
	w.Header().Set("X-Duration-Milliseconds", fmt.Sprintf("%d", elapsed))
	fmt.Fprintf(w, out)
}

func main() {
	flag.StringVar(&fndir, "d", ".", "function directory")
	flag.Parse()

	http.HandleFunc("/", handle)
	http.HandleFunc("/pipe/", handlePipe)
	http.ListenAndServe(":8080", nil)
}
