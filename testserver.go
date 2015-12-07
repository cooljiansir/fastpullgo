package main

import(
	"os"
	"fmt"
	"github.com/cooljiansir/fastpush/server"
)

func main(){
	if len(os.Args) < 2{
		fmt.Println("format: test path")
	}
	path := os.Args[1]
	server.Scan(path)
}
