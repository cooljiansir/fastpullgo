package main

import(
	"os"
	"fmt"
	"io"
)


func main(){
	if len(os.Args) <=1 {
		fmt.Println("format: input")
		return
	}
	filestr := os.Args[1]
	file,err := os.Open(filestr)
	if err != nil{
		panic(err)
	}
	b := make([]byte,20,20)
	for{
		_,err := file.Read(b)
		if err == io.EOF{
			break
		}else if err != nil{
			panic(err)
		}
		fmt.Printf("[%x]\n",b)
	}
}
