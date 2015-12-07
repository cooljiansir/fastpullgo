package server

import (
	"io"
	"path/filepath"
	"os"
	"fmt"
	"errors"
	. "github.com/cooljiansir/fastpush/spliter"
)

type block struct {
	filepath string
	off	int
	length	int
}

var blockMap map[[HashSize]byte]block

func MapFile(mmap map[[HashSize]byte]block,filepath string){
	fmt.Println("scan",filepath)
	file,err := os.Open(filepath)
	if err != nil{
		panic(err)
	}
	s := NewSpliter(file,4*1024)
	blks := make([]Block,1,1)
	for{
		_,err := s.Read(blks)
		if err == io.EOF{
			break
		}else if err != nil{
			panic(err)
		}
		mmap[blks[0].Hash()] = block{
			filepath:filepath,
			off:blks[0].Offset(),
			length:blks[0].Length(),
		}
	}
}

func Scan(path string){
	if blockMap == nil{
		blockMap = make(map[[HashSize]byte]block)
	}
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
                if ( f == nil ) {return err}
                if f.IsDir() {return nil}
                MapFile(blockMap,path)
                return nil
        })
        if err != nil {
                fmt.Printf("filepath.Walk() returned %v\n", err)
        }
}

func ReadHelper(r io.Reader,b []byte)(int,error){
	if len(b) == 0{
		return 0,nil
	}
	readed := 0
	for{
		n,err := r.Read(b[readed:])
		if err != nil && err != io.EOF{
			return readed,err
		}
		if err == io.EOF{
			break
		}
		readed += n
		if readed == len(b) {
			return len(b),nil
		}
	}
	if readed == 0{
		return 0,io.EOF
	}
	return readed,nil
}

//IdxReader read the hash to 0111001
//1 means a blocks already exists,0 means not
type IdxReader struct{
	r io.Reader
}

func NewIdxReader(r io.Reader)*IdxReader{
	return &IdxReader{
		r:r,
	}
}


func (r *IdxReader)Read(b []byte)(int,error){
	if len(b) == 0 {
		return 0,nil
	}
	readed := 0
	buf := [HashSize]byte{}
	for {
		n,err := ReadHelper(r.r,buf[:])
		if n != HashSize{
			return readed,errors.New("fastpush IdxReader error: size % HashSize != 0")
		}
		if err == io.EOF {
			break
		}
		_,find := blockMap[buf]
		if find {
			b[readed] = '1'
		}else{
			b[readed] = '0'
		}
		readed ++ 
		if readed == len(b){
			break
		}
	}
	if readed == 0{
		return 0,io.EOF
	}
	return readed,nil
}

//CntReader read the content data(part)
//and rebuild the whole data
//[------hash-------][length][----data of a block---]
//[------hash-------][000000]
type CntReader struct{
	r io.Reader
}


func NewCntReader(r io.Reader)*CntReader{
	return &CntReader{
		r:r,
	}
}
