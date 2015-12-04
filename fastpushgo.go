package fastpush

import (
	"bufio"
	hasher "crypto/sha1"
	"io"
	"math"
	"os"
)

const ModeFixed = 0
const HashSize = hasher.Size
const BlockSize = 1024 * 4

type HashBlockMap map[[HashSize]byte]Block

type Block struct {
	filename string
	offset   int
	length   int
}

func (b *Block) Length() int {
	return b.length
}
func (b *Block) Offset() int {
	return b.offset
}

func MapFile(m map[[HashSize]byte]Block, file string) {
	ifile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(ifile)
	defer ifile.Close()
	buf := make([]byte, BlockSize, BlockSize)
	readed := 0
	for {
		len, err := r.Read(buf)
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
		sum := hasher.Sum(buf[:len])
		m[sum] = Block{
			filename: file,
			offset:   readed,
			length:   len,
		}
		readed += len
	}
}
func MapFile2(m HashBlockMap, file string) {
	maxSize := BlockSize
	var h uint32     // rolling hash for finding fragment boundaries
	var c1 byte      // last byte
	var o1 [256]byte // order 1 context -> predicted byte
	fragment := math.Log2(float64(maxSize) / (64 * 64))
	mh := math.Exp2(22 - fragment)
	maxFragment := int(maxSize)
	minFragment := int(maxSize / 64)
	maxHash := uint32(mh)

	ifile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(ifile)
	defer ifile.Close()
	buf := make([]byte, 1, 1)
	readed := 0
	len := 0
	blockbuf := []byte{}
	for {
		readlen, err := r.Read(buf)
		if err == io.EOF || readlen == 0 {
			return
		}
		if err != nil {
			panic(err)
		}
		c := buf[0]
		if c == o1[c1] {
			h = (h + uint32(c) + 1) * 314159265
		} else {
			h = (h + uint32(c) + 1) * 271828182
		}
		blockbuf = append(blockbuf, c)
		o1[c1] = c
		c1 = c
		readed++
		len++

		// At a break point? Send it off!
		if (len >= minFragment && h < maxHash) || len >= maxFragment {
			sum := hasher.Sum(blockbuf)
			m[sum] = Block{
				filename: file,
				offset:   readed - len,
				length:   len,
			}
			len = 0
			blockbuf = []byte{}
			h = 0
			c1 = 0
		}
	}
}
