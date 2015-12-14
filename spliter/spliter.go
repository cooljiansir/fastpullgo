package spliter

import (
	"bufio"
	hasher "crypto/sha1"
	"io"
	"math"
)

//HashSize is the hash size
const HashSize = hasher.Size

//Block contains information about a block
type Block struct{
	data []byte
	hash   [HashSize]byte
	offset int64
}

//Hash  hash
func (b *Block) Hash() [HashSize]byte {
	return b.hash
}

//Offset offset
func (b *Block) Offset() int64 {
	return b.offset
}

//data
func (b *Block) Data()[]byte{
	return b.data
}

//Spliter split into Blocks
type Spliter struct {
	h uint32     // rolling hash for finding fragment boundaries
        c1 byte      // last byte
        o1 [256]byte // order 1 context -> predicted byte
        maxFragment int
        minFragment int
        maxHash uint32
	reader	*bufio.Reader
	readed	int64
}

func NewSpliter(r io.Reader,maxSize uint) *Spliter{
	fragment := math.Log2(float64(maxSize) / (64 * 64))
	mh := math.Exp2(22 - fragment)
	return &Spliter{
		maxFragment: int(maxSize),
		minFragment: int(maxSize / 64),
		maxHash:     uint32(mh),
		reader:	     bufio.NewReader(r),
		readed:	     0,
	}
}

func (s *Spliter)Read(b []Block)(int,error){
	countb := 0
	c1 := s.c1
	h := s.h

	if len(b) == 0{
		return 0,nil
	}
	blockBuf := []byte{}
	for{
		c,err := s.reader.ReadByte()
		if err == io.EOF && c == 0 {
			break
		}else if err != nil{
			s.h = h
			s.c1 = c1
			return countb,err
		}
		if c == s.o1[c1] {
			h = (h + uint32(c) + 1) * 314159265
		} else {
			h = (h + uint32(c) + 1) * 271828182
		}
		s.o1[c1] = c
		c1 = c
		s.readed ++
		blockBuf = append(blockBuf,c)
		if (len(blockBuf) >= s.minFragment && h < s.maxHash) || len(blockBuf) >= s.maxFragment {
			sum := hasher.Sum(blockBuf)
                        nblk := Block{
				data:blockBuf,
                                hash:   sum,
                                offset: s.readed - int64(len(blockBuf)),
                        }
			b[countb] = nblk
			countb ++
			blockBuf = []byte{}
			h = 0
			c1 = 0
			if countb >= len(b){
				break
			}
		}
	}
	if len(blockBuf) > 0 {
		sum := hasher.Sum(blockBuf)
                nblk := Block{  
			data:blockBuf,
                        hash:   sum,
                        offset: s.readed - int64(len(blockBuf)),
                }
		b[countb] = nblk
		countb ++
	}
	s.h = h
	s.c1 = c1
	if countb == 0 {
		return 0,io.EOF
	}
	return countb,nil
}

