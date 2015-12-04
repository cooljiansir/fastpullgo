package fastpush


import (
	"io"
	hasher "crypto/sha1"
	"bufio"
	"math"

	"github.com/cooljiansir/fastpush"
)

const HashSize = hasher.Size

type Block struct{
	hash   [HashSize]byte
        offset int
        length int
}
func (b *Block)Hash()[HashSize]byte{
	return b.hash
}
func (b *Block)Length()int{
	return b.length
}
func (b *Block)Offset()int{
	return b.offset;
}


//split into dynamic blocks
//maxSize is the max size of a block,min size of a block is maxSize/64
//maxCount is the max block count,0 means no limit
func Split(br *bufio.Reader,maxSize int,maxCount int)[]Block{
	var h uint32            // rolling hash for finding fragment boundaries
        var c1 byte             // last byte
        var o1 [256]byte        // order 1 context -> predicted byte
        fragment := math.Log2(float64(maxSize) / (64 * 64))
        mh := math.Exp2(22 - fragment)
        maxFragment := int(maxSize)
        minFragment := int(maxSize / 64)
        maxHash := uint32(mh)

	readed := 0
        bufLen := 0
        blockBuf :=[]byte{}

	res := []Block{}
	
	for {
                c,err := br.ReadByte()
                if err == io.EOF {
			break
		}
                if err != nil{
			panic(err)
		}
                if c == o1[c1] {
                        h = (h + uint32(c) + 1) * 314159265
                } else {
                        h = (h + uint32(c) + 1) * 271828182
                } 
                blockBuf = append(blockBuf,c)
                o1[c1] = c
                c1 = c
                readed ++
                bufLen ++ 
                
                // At a break point
                if (bufLen >= minFragment && h < maxHash) || bufLen >= maxFragment {
                        sum := hasher.Sum(blockBuf)
                        nblk := Block{
				hash:sum,
                                offset:readed-bufLen,
                                length:bufLen,
                        }
			res = append(res,nblk)
                        bufLen = 0 
                        blockBuf = []byte{}
                        h = 0 
                        c1 = 0
			if maxCount >= 0 && len(res) >= maxCount{
				break
			}
                }
        }
	if len(blockBuf) > 0 {
                sum := hasher.Sum(blockBuf)
                nblk := Block{
			hash:sum,
                        offset:readed-bufLen,
                	length:bufLen,
                }
		res = append(res,nblk)
	}
	return res
}
