package main

import(
	"fmt"
	"time"
	hasher "crypto/sha1"
)


type myhash struct{
	bucket []map[string]string
}

func NewMap()*myhash{
	mm := make([]map[string]string,1 << 24,1 << 24)
	return &myhash{
		bucket:mm,
	}
}

func (m *myhash)find(k string)(string,bool){
	kb := []byte(k)
	h := hasher.Sum(kb)

	index := uint(h[0]) + uint(h[1]) << 8 + uint(h[2]) << 16
	if m.bucket[index] == nil{
		return "",false
	}
	v,find := m.bucket[index][k]
	return v,find
}

func (m *myhash)insert(k string,v string){
	kb := []byte(k)
	h := hasher.Sum(kb)

	index := uint(h[0]) + uint(h[1]) << 8 + uint(h[2]) << 16
	if m.bucket[index] == nil{
		m.bucket[index] = make(map[string]string)
	}
	m.bucket[index][k] = v
}

func testmytree(n int){
        t := NewMap()
        fmt.Println("mytree insert begin")
        start := time.Now()
        for i:=0;i<n;i++{
                index := fmt.Sprintf("%d",i)
                t.insert(index,index)
        }
        end := time.Now()
        st := end.Sub(start).Nanoseconds()/1000000
        fmt.Println("mytree insert end",st)

        fmt.Println("mytree lookup begin")
        start = time.Now()
        for i:=0;i<n;i++{
                index := fmt.Sprintf("%d",i)
                _,find := t.find(index)
                if !find {
                        fmt.Println("not found: ",index)
                }
        }
        end = time.Now()
        st = end.Sub(start).Nanoseconds()/1000000
        fmt.Println("mytree lookup end",st)
}

func testgolang(n int){
        m := make(map[string]string,n)
        fmt.Println("golang map insert begin")
        start := time.Now()
        for i:=0;i<n;i++{
                index := fmt.Sprintf("%d",i)
                m[index] = index
        }
        end := time.Now()
        t := end.Sub(start).Nanoseconds()/1000000
        fmt.Println("golang map insert end",t)

        fmt.Println("golang map lookup begin")
        start = time.Now()
        for i:=0;i<n;i++{
                index := fmt.Sprintf("%d",i)
                _,find := m[index]
                if !find {
                        fmt.Println("not found: ",index)
                }
        }
        end = time.Now()
        t = end.Sub(start).Nanoseconds()/1000000
        fmt.Println("golang map lookup end",t)
}

func testtime(n int){
	testmytree(n)
	//testgolang(n)
}

func main(){
	testtime(10000000)	
}
