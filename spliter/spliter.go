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
	hash   [HashSize]byte
	offset int
	length int
}

//Hash  hash
func (b *Block) Hash() [HashSize]byte {
	return b.hash
}

//Length length
func (b *Block) Length() int {
	return b.length
}

//Offset offset
func (b *Block) Offset() int {
	return b.offset
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
	readed	int
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
                                hash:   sum,
                                offset: s.readed - len(blockBuf),
                                length: len(blockBuf),
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
                        hash:   sum,
                        offset: s.readed - len(blockBuf),
                        length: len(blockBuf),
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


//Split into dynamic blocks
//maxSize is the max size of a block,min size of a block is maxSize/64
//maxCount is the max block count,0 means no limit
func Split(br *bufio.Reader, maxSize int, maxCount int) []Block {
	var h uint32     // rolling hash for finding fragment boundaries
	var c1 byte      // last byte
	var o1 [256]byte // order 1 context -> predicted byte
	fragment := math.Log2(float64(maxSize) / (64 * 64))
	mh := math.Exp2(22 - fragment)
	maxFragment := int(maxSize)
	minFragment := int(maxSize / 64)
	maxHash := uint32(mh)

	readed := 0
	bufLen := 0
	blockBuf := []byte{}

	res := []Block{}

	for {
		c, err := br.ReadByte()
		if err == io.EOF && c == 0{
			break
		}
		if err != nil {
			panic(err)
		}
		if c == o1[c1] {
			h = (h + uint32(c) + 1) * 314159265
		} else {
			h = (h + uint32(c) + 1) * 271828182
		}
		blockBuf = append(blockBuf, c)
		o1[c1] = c
		c1 = c
		readed++
		bufLen++

		// At a break point
		if (bufLen >= minFragment && h < maxHash) || bufLen >= maxFragment {
			sum := hasher.Sum(blockBuf)
			nblk := Block{
				hash:   sum,
				offset: readed - bufLen,
				length: bufLen,
			}
			res = append(res, nblk)
			bufLen = 0
			blockBuf = []byte{}
			h = 0
			c1 = 0
			if maxCount > 0 && len(res) >= maxCount {
				break
			}
		}
	}
	if len(blockBuf) > 0 {
		sum := hasher.Sum(blockBuf)
		nblk := Block{
			hash:   sum,
			offset: readed - bufLen,
			length: bufLen,
		}
		res = append(res, nblk)
	}
	return res
}
