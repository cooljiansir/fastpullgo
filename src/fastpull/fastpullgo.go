package fastpull


import (
	"io"
	hasher "crypto/sha1"
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
