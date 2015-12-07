package main

import (
	"github.com/cooljiansir/fastpush/client"
	"os"
	"fmt"
	"io"
)



func main(){	
	if len(os.Args) < 3{
		fmt.Println("Input format : test input output")
		return
	}
	filestr := os.Args[1]
	fileoutstr := os.Args[2]
	file,err := os.Open(filestr)
	defer file.Close()
	if err != nil{
		panic(err)
	}
	ofile,err := os.Create(fileoutstr)
	if err != nil{
		panic(err)
	}
	defer ofile.Close()
	r := client.NewIdxReader(file)
	/*for {
		b := make([]byte,1024,1024)
		_,err := r.Read(b)
		if err == io.EOF{
			break
		}else if err != nil{
			panic(err)
		}
		fmt.Printf("%x\n",b)
	}
	return*/
	_,err = io.Copy(ofile,r)
	if err != nil{
		panic (err)
	}
}
