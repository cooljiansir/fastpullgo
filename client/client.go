package client


import(
	"io"
	. "github.com/cooljiansir/fastpush/spliter"
)

//IdxReader read out all the hash of a file
type IdxReader struct {
	r io.Reader
	cur []byte
	s *Spliter
}

func NewIdxReader(r io.Reader)* IdxReader{
	return &IdxReader{
		r:r,
		cur:[]byte{},
		s:NewSpliter(r,1024*4),
	}
}

func (r *IdxReader)Read(b []byte)(int,error){
	readb := 0
	blks := make([]Block,1,1)
	for{
		if len(r.cur) == 0{
			_,err := r.s.Read(blks)
			if err == io.EOF{
				break
			}else if err != nil{
				return readb,err
			}
			hash := blks[0].Hash()
			r.cur = hash[:]
		}
		n := copy(b[readb:],r.cur)
		readb += n
		r.cur = r.cur[n:]
		if readb >= len(b){
			break
		}
	}
	if readb == 0{
		return 0,io.EOF
	}
	return readb,nil
}


//CntReader Read part of the data of content
//
type CntReader struct {
	r io.Reader
	idxReader *IdxReader
}


func NewCntReader(r io.Reader,ir *IdxReader)*CntReader{
	return &CntReader{
		r:r,
		idxReader:ir,
	}
}
