package fastpush

import (
	"io"
	"github.com/cooljiansir/fastpush"
)


var blockMap map[[HashSize]byte]Block

func Scan(){
	if blockMap == nil{
		blockMap = make(map[[HashSize]byte]Block)
	}
	
}

func readHelper(r io.Reader,b []byte)(int,error){
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

func NewIdxReader(r io.Reader)*IdexReader{
	return &IdxReader{
		r:r,
	}
}


func (r *IdxReader)Read(b []byte)(int,error){
	
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
