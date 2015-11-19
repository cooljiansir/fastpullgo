package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"fastpull"
)

var hashmap map[[fastpull.HashSize]byte]fastpull.Block

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
func hashHandle(w http.ResponseWriter,r *http.Request){
	fmt.Println("A request is comming")
	if hashmap != nil{
		for{
			buf := make([]byte,fastpull.HashSize,fastpull.HashSize)
			len,err := r.Body.Read(buf)
			if err == io.EOF{
				return
			}
			if err != nil{
				io.WriteString(w,"err")
				return
			}
			if len != fastpull.HashSize{
				io.WriteString(w,"size not HashSize")
				return
			}
			h := [fastpull.HashSize]byte{}
			copy(h[:],buf)
			_,find := hashmap[h]
			if !find{
				io.WriteString(w,"0")
			}else{
				io.WriteString(w,"1")
			}
		}
	}
}
func main(){
	hashmap = make(map[[fastpull.HashSize]byte]fastpull.Block)
	for i,name := range os.Args {
		if i>0{
			fmt.Println("scan ",name)
			fastpull.MapFile(hashmap,name)
			fmt.Println("end")
		}
	}

	for h,_ := range hashmap{
		fmt.Printf("%x\n",h)
	}
	http.HandleFunc("/upload",uploadHandle)
	http.HandleFunc("/hash",hashHandle)
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		panic(err)
	}
}
