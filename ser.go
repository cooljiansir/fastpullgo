package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)


func uploadHandle(w http.ResponseWriter,r *http.Request){
	if r.Method == "GET" {
		io.WriteString(w,"hello!")
	}else {
		file,_,err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return 
		}
		defer file.Close()
		fout,err := os.Create("out")
		if err != nil {
			fmt.Println(err)
			return 
		}
		defer fout.Close()
		_,err = io.Copy(fout,file)
		if err != nil {
			fmt.Println(err)
			return 
		}
		io.WriteString(w,"upload succeed")
	}
}
func main(){
	http.HandleFunc("/upload",uploadHandle)
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("succeed")
}
