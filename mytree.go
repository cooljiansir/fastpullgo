package main


import (
	"fmt"
	"time"
        hasher "crypto/sha1"
)

type kv struct{
	key string
	value string
}

type node struct{
	next [TREEN]*node
	kv *kv
}


const POOLSIZE = 1024*50
const BIT = 2


const TREEN = 1<<BIT
const FILTER = TREEN - 1

type poolnode struct{
	pool [POOLSIZE]node
	next *poolnode
}

type mytree struct{
	root *node
	countnode int
	poolroot *poolnode
	poolrear *poolnode
}
func NewMap()*mytree{
	nn := &node{
		next:[TREEN]*node{},
		kv:nil,
	}
	pp := &poolnode{
		pool:[POOLSIZE]node{},
		next:nil,
	}
	return &mytree{
		root:nn,
		countnode:0,
		poolroot:pp,
		poolrear:pp,
	}
}

func (t *mytree)NewNode()*node{
	if t.countnode == POOLSIZE{
		t.poolrear.next = &poolnode{
			pool:[POOLSIZE]node{},
			next:nil,
		}
		t.poolrear = t.poolrear.next
		t.countnode = 0
	}
	nn := &t.poolrear.pool[t.countnode]
	t.countnode ++
	return nn
}
var maxdepth uint
func (t *mytree)find(k string)(string,bool){
	p := t.root
	
	kb := []byte(k)
	h := hasher.Sum(kb)
	bit := uint(0)
	var idn uint
	for ;p!=nil;{
		if p.kv != nil{
			break
		}
		idn = ((uint(h[(bit)>>3])) >> (bit&7)) & FILTER
		p = p.next[idn]
		bit += BIT
	}
	if bit > maxdepth{maxdepth=bit}
	if p != nil && p.kv.key == k{
		return p.kv.value,true
	}
	return "",false
}

func (t *mytree)insert(pair kv){
	p := t.root
	pf := p
	
	kb := []byte(pair.key)
        h := hasher.Sum(kb)
        bit := uint(0)
        var idn uint

        for ;p!=nil;{   
                if p.kv != nil{
                        break
                }
		idn = ((uint(h[(bit)>>3])) >> (bit&7) ) & FILTER
		pf = p
                p = p.next[idn]
		bit += BIT
        }
        if p != nil {
		if p.kv.key == pair.key {
			p.kv.value = pair.value
			return
		}
		kvp := p.kv
		p.kv = nil
		kvpb := []byte(kvp.key)
		hash0 := hasher.Sum(kvpb)
		var idn0 uint
		for {
			idn0 = ((uint(hash0[(bit)>>3])) >> (bit&7) ) & FILTER
			idn = ((uint(h[(bit)>>3])) >> (bit&7) ) & FILTER
			if idn == idn0{
				nn := t.NewNode()
				p.next[idn] = nn
				p = nn
			}else{
				nn := t.NewNode()
				nn.kv = &pair
				nn0 := t.NewNode()
				nn0.kv = kvp
				p.next[idn0] = nn0
				p.next[idn] = nn
				return
			}
			bit += BIT
		}
	}else{
		nn := t.NewNode()
		nn.kv = &pair
		pf.next[idn] = nn
	}
}

func testempty(n int){
        fmt.Println("loop begin")
        start := time.Now()
        for i:=0;i<n;i++{
               fmt.Sprintf("hello world %d",i)	
        }
        end := time.Now()
        st := end.Sub(start).Nanoseconds()/1000000
        fmt.Println("loop end",st)
}
func testmytree(n int){
	t := NewMap()
	fmt.Println("mytree insert begin")
	start := time.Now()	
	for i:=0;i<n;i++{
		index := fmt.Sprintf("%d",i)
		t.insert(kv{
			key:index,
			value:index,
		})
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
	m := make(map[string]string)
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
	testempty(n)
	testmytree(n)
	fmt.Println("maxdepth:",maxdepth)
	testgolang(n)
}


func main(){
	testtime(10000000)	
}
