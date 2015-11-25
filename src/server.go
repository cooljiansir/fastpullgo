package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"fastpull"
	"bytes"
)

var hashmap map[[fastpull.HashSize]byte]fastpull.Block


func ReadFull(r io.Reader,b []byte)(int,error){
	buf := make([]byte,1,1)
	readed := 0
	for{
		if readed >= len(b){break}
		_,err := r.Read(buf)
		if err == io.EOF{
			if readed ==0 {
				return 0,err
			}
			return readed,nil
		}
		if err != nil{
			return readed,nil
		}
		b[readed] = buf[0]
		readed ++
	}
	return readed,nil
}
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
	str := bytes.NewBufferString("")
	if hashmap != nil{
		for{
			buf := make([]byte,fastpull.HashSize,fastpull.HashSize)
			len,err := ReadFull(r.Body,buf)
			if err == io.EOF{
				break
			}
			if err != nil{
				str.WriteString(err.Error())
				break
			}
			if len != fastpull.HashSize{
				break
			}
			h := [fastpull.HashSize]byte{}
			copy(h[:],buf)
			_,find := hashmap[h]
			if !find{
				str.WriteString("0")
			}else{
				str.WriteString("1")
			}
		}
		io.Copy(w,str)
	}
}
func main(){
	hashmap = make(map[[fastpull.HashSize]byte]fastpull.Block)
	isFixed := true 
	for _,arg := range os.Args{
		if arg == "-d"{
			isFixed = false
			break
		}
	}
	for i,name := range os.Args {
		if i>0{
			fmt.Println("scan ",name)
			if name == "-d"{continue}
			if isFixed{
				fastpull.MapFile(hashmap,name)
			}else{
				fastpull.MapFile2(hashmap,name)
			}
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
