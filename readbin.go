package main

import(
	"os"
	"fmt"
	"bufio"
	"github.com/cooljiansir/fastpush/spliter"
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
	r := bufio.NewReader(file)
	blks := spliter.Split(r,4*1024,0)
	for _,b := range blks{
		fmt.Printf("%d[%x]\n",b.Length(),b.Hash())
	}
}
