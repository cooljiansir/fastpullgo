package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/cooljiansir/fastpush/server"
)

func hashHandle(w http.ResponseWriter,r *http.Request){
	fmt.Println("A request is comming")
	reader := server.NewIdxReader(r.Body)
	io.Copy(w,reader)
}

func main(){
	if len(os.Args) <2{
		fmt.Println("format: test path")
		return
	}
	path := os.Args[1]
	server.Scan(path)
	http.HandleFunc("/hash",hashHandle)
	fmt.Println("listening...")
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		panic(err)
	}
}
