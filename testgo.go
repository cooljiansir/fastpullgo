package main

import(
	"fmt"
)


func test(){
	a := []int{1,2,3}
	b := a[3:]
	fmt.Println(len(b))
}

type metadata struct {
	filename string
	offset uint64
	length uint32
}

func teststruct_(m *metadata){
	m.filename = "after set value"
}

func teststruct(){
	m := &metadata{
		filename:"before",
	}
	teststruct_(m)
	fmt.Println(m.filename)
}


func main(){
	teststruct()
}
