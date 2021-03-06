// SimpleFileServer project main.go
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kolonse/KolonseWeb"
	"github.com/kolonse/KolonseWeb/HttpLib"
	"github.com/kolonse/KolonseWeb/Type"
)

var port = flag.Int("-p", 54321, "-p=<port> default=54321")
var ip = flag.String("-h", "0.0.0.0", "-h=<ip> default=0.0.0.0")

func response(res *HttpLib.Response, code string, desc string) {
	res.Header().Set("Code", code)
	res.Header().Set("Message", desc)
}

func main() {
	flag.Parse()
	KolonseWeb.DefaultApp.Post("/upload", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		dst := req.URL.Query().Get("dst")
		baseDir := filepath.Dir(dst)
		os.MkdirAll(baseDir, os.ModePerm)
		file, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			//			res.End(Response(-1, err.Error()))
			response(res, "-1", err.Error())
			return
		}
		response(res, "0", "")
		defer req.Body.Close()
		defer file.Close()
		r := bufio.NewReader(req.Body)
		w := bufio.NewWriter(file)
		_, err = io.Copy(w, r)
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
		err = w.Flush()
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
	})
	KolonseWeb.DefaultApp.Post("/download", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		src := req.URL.Query().Get("src")
		file, err := os.OpenFile(src, os.O_RDONLY, 0666)
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
		response(res, "0", "")
		defer req.Body.Close()
		defer file.Close()
		r := bufio.NewReader(file)
		w := bufio.NewWriter(res)
		_, err = io.Copy(w, r)
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
		err = w.Flush()
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
	})
	KolonseWeb.DefaultApp.Post("/cmd", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		cmd := req.URL.Query().Get("cmd")
		arg := req.URL.Query().Get("arg")
		var argArr []string
		err := json.Unmarshal([]byte(arg), &argArr)
		w := bufio.NewWriter(res)
		response(res, "0", "")
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
		c := exec.Command(cmd, argArr...)
		c.Stdin = os.Stdin
		oor, e1 := c.StdoutPipe()
		ooe, e2 := c.StderrPipe()
		err = c.Start()
		if err != nil {
			response(res, "-1", err.Error())
			return
		}
		if e1 == nil {
			io.Copy(w, oor)
			w.Flush()
		} else {
			KolonseWeb.Warning(cmd, arg, "io", e1.Error())
		}
		if e2 == nil {
			io.Copy(w, ooe)
			w.Flush()
		} else {
			KolonseWeb.Warning(cmd, arg, "io", e2.Error())
		}
		if err := c.Wait(); err != nil {
			response(res, "-1", err.Error())
		}
	})
	KolonseWeb.DefaultApp.Listen(*ip, *port)
}
