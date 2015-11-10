package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)


func Upload(url,file string) (err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	f,err := os.Open(file)
	if err != nil {
		panic(err)
	}
	fw,err := w.CreateFormFile("file",file)
	if err != nil {
		panic(err)
	}
	_,err = io.Copy(fw,f)
	if err != nil {
		panic(err)
	}
	w.Close()
	req,err := http.NewRequest("POST",url,&b)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type",w.FormDataContentType())
	client := &http.Client{}
	res,err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status:%s",res.Status)
	}
	return err
}

func main() {
	Upload("http://10.10.19.104:8080/upload","test")
}
