package main


import(
	"os"
	"fmt"
)

func main(){
	if len(os.Args)<=1{
		fmt.Println("Please input file")
		return
	}else{
		buf := make([]byte,1024,1024)
		file,err := os.Open(os.Args[1])
		if err != nil{
			panic(err)
		}
		defer file.Close()
		_,err = file.Read(buf)
		if err != nil{
			panic(err)
		}
	}
}
