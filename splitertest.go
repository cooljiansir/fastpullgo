package main

import (
	"github.com/cooljiansir/fastpush/spliter"
	"os"
	"fmt"
	"io"
)

func main(){
	filestr := os.Args[1]
	file,err := os.Open(filestr)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	s := spliter.NewSpliter(file,4*1024)
	for{
		b := make([]spliter.Block,10,10)
		_,err :=s.Read(b)
		if err == io.EOF{
			break
		}else if err != nil{
			panic(err)
		}
		for _,blk := range b{
			fmt.Printf("%d [%x]\n",blk.Length(),blk.Hash())
		}
	}
}
