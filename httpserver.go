package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/cooljiansir/fastpush/server"
)

func hashHandle(w http.ResponseWriter,r *http.Request){
	fmt.Println("A hash request is comming")
	reader := server.NewIdxReader(r.Body)
	io.Copy(w,reader)
}
func fileHandle(w http.ResponseWriter,r *http.Request){
	fmt.Println("A file request is comming")
	reader := server.NewCntReader(r.Body)
	ofile,err := os.Create("upload")
	if err != nil{
		panic(err)
	}
	defer ofile.Close()
	io.Copy(ofile,reader)
	reader.Close()
}

func main(){
	http.HandleFunc("/hash",hashHandle)
	http.HandleFunc("/file",fileHandle)
	fmt.Println("listening...")
	server.Start("/var/fastpush/")
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		panic(err)
	}
}
