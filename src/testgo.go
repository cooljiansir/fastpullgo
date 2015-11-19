package main

import(
	"fmt"
	"os"
)

func main(){
	file,err := os.Open("test")
	if err != nil{
		panic(err)
	}	
	buf := make([]byte,1024,1024)
	n,err := file.Read(buf)
	fmt.Println("read byte:",n," buf size ",len(buf))
}
