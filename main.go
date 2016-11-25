// SimpleFileServer project main.go
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"github.com/kolonse/KolonseWeb"
	"github.com/kolonse/KolonseWeb/HttpLib"
	"github.com/kolonse/KolonseWeb/Type"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var Port = flag.Int("-p", 54321, "-p=<port> default=54321")

func Response(res *HttpLib.Response, code string, desc string) {
	//	buf, _ := json.Marshal(map[string]interface{}{
	//		"Code":    code,
	//		"Message": desc,
	//	})
	res.Header().Set("Code", code)
	res.Header().Set("Message", desc)
	//	res.End()
}

func main() {
	flag.Parse()
	KolonseWeb.DefaultApp.Post("/upload", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		dst := req.URL.Query().Get("dst")
		baseDir := filepath.Dir(dst)
		os.MkdirAll(baseDir, 666)
		file, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			//			res.End(Response(-1, err.Error()))
			Response(res, "-1", err.Error())
			return
		}
		defer req.Body.Close()
		defer file.Close()
		r := bufio.NewReader(req.Body)
		w := bufio.NewWriter(file)
		_, err = io.Copy(w, r)
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		err = w.Flush()
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		Response(res, "0", "")
	})
	KolonseWeb.DefaultApp.Post("/download", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {
		src := req.URL.Query().Get("src")
		file, err := os.OpenFile(src, os.O_RDONLY, 0666)
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		defer req.Body.Close()
		defer file.Close()
		r := bufio.NewReader(file)
		w := bufio.NewWriter(res)
		_, err = io.Copy(w, r)
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		err = w.Flush()
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		Response(res, "0", "")
	})
	KolonseWeb.DefaultApp.Post("/cmd", func(req *HttpLib.Request, res *HttpLib.Response, next Type.Next) {

		cmd := req.URL.Query().Get("cmd")
		arg := req.URL.Query().Get("arg")
		var argArr []string
		err := json.Unmarshal([]byte(arg), &argArr)
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		w := bufio.NewWriter(res)
		c := exec.Command(cmd, argArr...)
		c.Stdin = os.Stdin
		c.Stdout = w
		c.Stderr = w
		err = c.Run()
		if err != nil {
			Response(res, "-1", err.Error())
			return
		}
		Response(res, "0", "")
		w.Flush()
	})
	KolonseWeb.DefaultApp.Listen("0.0.0.0", *Port)
}