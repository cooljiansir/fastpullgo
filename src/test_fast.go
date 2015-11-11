package main

import(
	"os"
	"fastpull"
	"io"
)

func test(){
	infile,err := os.Open("test")
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	reader := fastpull.NewHashReader(infile,fastpull.ModeFixed)
	oufile,err := os.Create("out")
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	io.Copy(oufile,reader)
}


func main(){
	test()
}
