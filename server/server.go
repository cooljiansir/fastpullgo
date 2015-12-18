package server

import (
	"io"
	"path/filepath"
	"os"
	"fmt"
	"errors"
	"encoding/binary"
	"bufio"
	. "github.com/cooljiansir/fastpush/spliter"
)

type block struct {
	filepath string
	off	int64
	length	int64
}

var blockMap map[[HashSize]byte]block
var fileOpenSeeker FileOpenSeeker
var walker Walker


func SetFileOpenSeeker(o FileOpenSeeker){
	fileOpenSeeker = o
}

func SetWalker(w Walker){
	walker = w
}


func mapFile(mmap map[[HashSize]byte]block,filepath string){
	fmt.Println("scan",filepath)
	if fileOpenSeeker == nil { return }
	file,err := fileOpenSeeker.OpenSeek(filepath,0)
	if err != nil{
		panic(err)
	}
	s := NewSpliter(file,4*1024)
	blks := make([]Block,1,1)
	for{
		n,err := s.Read(blks)
		if err == io.EOF && n == 0{
			break
		}else if err != nil && err != io.EOF{
			panic(err)
		}
		mmap[blks[0].Hash()] = block{
			filepath:filepath,
			off:blks[0].Offset(),
			length:int64(len(blks[0].Data())),
		}
	}
}
type Walker interface{
	Walk(basepath string,f func(path string,isDir bool)error)error
}
type DefaultWalker struct{
}
func NewDefaultWalker()*DefaultWalker{
	return &DefaultWalker{}
}

func (w *DefaultWalker) Walk(basepath string,fc func(path string,isDir bool) error ) error {
	return filepath.Walk(basepath,func(path string, f os.FileInfo, err error) error {
                if ( f == nil ) {return err}
		return fc(path,f.IsDir())
        })
}
func Scan(path string){
	if walker == nil { return }
	if blockMap == nil{
		blockMap = make(map[[HashSize]byte]block)
	}
	err := walker.Walk(path, func(path string, isDir bool) error {
                if isDir {return nil}
                mapFile(blockMap,path)
                return nil
        })
        if err != nil {
                fmt.Printf("Walk returned %v\n", err)
        }
}

func ReadHelper(r io.Reader,b []byte)(int,error){
	if len(b) == 0{
		return 0,nil
	}
	readed := 0
	for{
		n,err := r.Read(b[readed:])
		if err == io.EOF && n == 0{
			break
		}else if err != nil && err != io.EOF{
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
		if err == io.EOF && n == 0{
			break
		}else if err != nil && err != io.EOF{
			return readed,err
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

// FileSeek returns an  io.ReadSeeker according to path
type FileOpenSeeker interface{
	OpenSeek(path string,off int64)(io.ReadCloser,error)
}

type DefaultFileOpenSeeker struct {
}

func NewDefaultFileOpenSeeker()* DefaultFileOpenSeeker{
	return &DefaultFileOpenSeeker{}
}

func (r *DefaultFileOpenSeeker)OpenSeek(path string,off int64)(io.ReadCloser,error){
	file,err := os.Open(path)
	if err != nil{
		return nil,err
	}
	_,err  = file.Seek(off,0)
	if err != nil{
		return nil,err
	}
	return file,nil
}


//CntReader read the content data(part)
//and rebuild the whole data
//[------hash-------][length][----data of a block---]
//[------hash-------][000000]
type CntReader struct{
	r *bufio.Reader
	cur []byte	//current reading
}


func NewCntReader(r io.Reader)*CntReader{
	return &CntReader{
		r:bufio.NewReader(r),
		cur:[]byte{},
	}
}

func (r *CntReader)Read(b []byte)(int,error){
	if fileOpenSeeker == nil{
		return 0,fmt.Errorf("file Opener is nil")
	}
	if len(b)==0{
		return 0,io.EOF
	}
	readed := 0
	hashbuf := [HashSize]byte{}
	for{
		if len(r.cur)==0{
			n,err := ReadHelper(r.r,hashbuf[:])
			if err == io.EOF && n == 0{
				break
			}else if err != nil && err != io.EOF{
				return readed,err
			}
			if n != HashSize {
				return readed,fmt.Errorf("read format error: size not HashSize")
			}
			//fmt.Printf("read hash [%x]\n",hashbuf)
			length, err := binary.ReadUvarint(r.r)
			if err != nil {
				return readed,err
			}
			//exists
			if length == 0{
				blk,find := blockMap[hashbuf]
				if !find {
					return readed,fmt.Errorf("hash not found")
				}
				filename := blk.filepath
				off := blk.off
				file,err := fileOpenSeeker.OpenSeek(filename,off)
				if err != nil{
					return readed,err
				}
				len := blk.length
				r.cur = make([]byte,len,len)
				n,err := ReadHelper(file,r.cur)
				if err != nil && err != io.EOF{
					return readed,err
				}
				if int64(n) != len{
					return readed,fmt.Errorf("read local file length wrong")
				}
				file.Close()
			}else{
				r.cur = make([]byte,length,length)
				n,err := ReadHelper(r.r,r.cur)
				if err != nil && err != io.EOF{
					return readed,err
				}
				if uint64(n) != length{
					return readed,fmt.Errorf("read net file length wrong")
				}
			}
		}
		n := copy(b[readed:],r.cur)
		readed += n
		r.cur = r.cur[n:]
		if readed >= len(b){
			break
		}
	}
	if readed == 0{
		return 0,io.EOF
	}
	return readed,nil
}


func init(){
	if blockMap == nil{
		blockMap = make(map[[HashSize]byte]block)
	}
	if fileOpenSeeker == nil{
		fileOpenSeeker = NewDefaultFileOpenSeeker()
	}
	if walker == nil{
		walker = NewDefaultWalker()
	}
}
