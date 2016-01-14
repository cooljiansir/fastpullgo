package main


import(
	"fmt"
)

func insertmap(m map[int]bool){
	for i:=0;i<1000000000;i++ {
		m[i] = true
		if i%1000==0{
			fmt.Printf("\r%d",i)
		}
	}
}
func main(){
	m := make(map[int]bool)
	fmt.Println("begin insert")
	insertmap(m)
	fmt.Println("insert end")
}
