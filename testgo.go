package main

import(
	"fmt"
)

func main(){
	a := []int{1,2,3}
	b := a[3:]
	fmt.Println(len(b))
}
