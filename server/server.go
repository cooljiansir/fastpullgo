package server

import (
	"io"
	"fmt"
	"errors"
	"encoding/binary"
	"bufio"
	. "github.com/cooljiansir/fastpush/spliter"
	. "github.com/cooljiansir/fastpush/fingerdb"
)


var fingerdb *FingerDB

const BASEPATH = "fastpush"
const DBFILE = "fastpush.db"

func init(){
	var err error
	fingerdb,err = NewFingerDB(DBFILE,BASEPATH)
	if err != nil{
		panic(err)
	}
	fmt.Println("fast push server started....")
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
	fmt.Println("NewIdxReader\n\n\n\n\n")

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
		find := false
		if fingerdb != nil{
			_,find = fingerdb.Find(buf)
		}
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
	r *bufio.Reader
	cur []byte		//current reading
	container *Container	//finger,block data container
}


func NewCntReader(r io.Reader)*CntReader{
	fmt.Println("NewCntReader\n\n\n\n\n")

	return &CntReader{
		r:bufio.NewReader(r),
		cur:[]byte{},
		container:fingerdb.NewContainer(),
	}
}

func (r *CntReader)Read(b []byte)(int,error){
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
				metadata,find := fingerdb.Find(hashbuf)
				if !find {
					return readed,fmt.Errorf("hash not found")
				}
				file,err := fingerdb.NewBlockReader(metadata)
				if err != nil{
					return readed,err
				}

				/*
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
				len := blk.length*/
				
				len := metadata.Length
				r.cur = make([]byte,len,len)
				n,err := ReadHelper(file,r.cur)
				if err != nil && err != io.EOF{
					return readed,err
				}
				if uint32(n) != len{
					return readed,fmt.Errorf("read local file length wrong")
				}
				file.Close()
			}else{
				r.cur = make([]byte,length,length)
				n,err := ReadHelper(r.r,r.cur)
				r.container.WriteBlock(hashbuf,r.cur)
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

func (r *CntReader)Close()error{
	return r.container.Close()	
}
