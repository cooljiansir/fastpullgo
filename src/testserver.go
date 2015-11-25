package main

import (
	"fmt"
	"io"
	"net/http"
)


func hashHandle(w http.ResponseWriter,r *http.Request){
	fmt.Println("A request is comming")
		for{
			buf := make([]byte,1,1)
			_,err := r.Body.Read(buf)
			if err == io.EOF{
				break
			}
			if err != nil{
				fmt.Println(err.Error())
				break
			}
			fmt.Println("read %s\n",string(buf))
		}
}
func main(){
	http.HandleFunc("/hash",hashHandle)
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		panic(err)
	}
}
