package main

import(
        "os"
	"fmt"
	"io"
        "github.com/cooljiansir/fastpush/client"
)


func main(){
        if len(os.Args) <3 {
                fmt.Println("format test infile outfile")
                return
        }       
        filestr := os.Args[1]
	fileoutstr := os.Args[2]
        file,err := os.Open(filestr)
        if err != nil{
                panic(err)
        }
	file2,err := os.Open(filestr)
	if err != nil{
		panic(err)
	}
	outfile,err := os.Create(fileoutstr)
	if err != nil{
		panic(err)
	}
        clt := client.NewClient(file,file2,"http://127.0.0.1:8080/hash")
        go clt.Start()
	if err != nil{
		panic(err)
	}
	io.Copy(outfile,clt)
}
