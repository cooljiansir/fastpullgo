package fastpull


import (
	"io"
	hasher "crypto/sha1"
	"os"
	"bufio"
)

const ModeFixed = 0
const HashSize = hasher.Size
const BlockSize = 1024

type HashReader struct{
        inReader io.Reader
        mode int
	cur []byte	//current block
}       
func NewHashReader(inr io.Reader,mode int)(*HashReader){
        w := &HashReader{
                inReader:inr,
                mode:mode,
        }
	return w;
}       
func (h *HashReader)Read(b []byte)(n int,err error){
	readed := 0
	for readed < len(b){
		if len(h.cur) == 0 {
			buf := make([]byte,BlockSize)
			len,err := h.inReader.Read(buf)
			buf = buf[:len]
			if err != nil{
				return readed,err
			}
			sum :=  hasher.Sum(buf)
			h.cur = sum[0:]
		}
		if len(h.cur) >0 {
			count := copy(b[readed:],h.cur)
			h.cur = h.cur[count:]
			readed += count
		}
	}
	return readed,nil
}
type HashBlockMap map[[HashSize]byte]Block


type Block struct{
        filename string
        offset int
         length int
}
func MapFile(m map[[HashSize]byte]Block,file string){
        ifile,err := os.Open(file)
        if err != nil{
                panic(err)
        }
        r := bufio.NewReader(ifile)
        defer ifile.Close()
        buf := make([]byte,BlockSize,BlockSize)
        readed := 0
        for {
                len,err := r.Read(buf)
		if err == io.EOF{return}
                if err!=nil{
                        panic(err)
                }
                sum := hasher.Sum(buf[:len])
                m[sum]=Block{
                        filename:file,
                        offset:readed,
                        length:len,
                }
                readed += len
        }
}
