package client


import(
	"io"
	"encoding/binary"
	"net/http"
	"fmt"
	"sync"
	. "github.com/cooljiansir/fastpush/spliter"
)

//IdxReader read out all the hash of a file
type IdxReader struct {
	r io.Reader
	cur []byte
	s *Spliter
	splited chan Block
	sclosed bool
	limit int	//max block count to read,if reached return io.EOF
	eof bool	//EOF of file (r EOF) NOT limit end
	readblks int	//blocks readed
	
	bytesRead int64	//bytes read
	rBytesRead int64//bytes read from r
}

func NewIdxReader(r io.Reader,splited chan Block,limit int)* IdxReader{
	return &IdxReader{
		r:r,
		cur:[]byte{},
		s:NewSpliter(r,BLKSIZE),
		splited:splited,
		limit:limit,
		eof:false,
		readblks:0,
		bytesRead:0,
		rBytesRead:0,
	}
}

func (r *IdxReader)Read(b []byte)(int,error){
	readb := 0
	blks := make([]Block,1,1)
	for{
		if len(r.cur) == 0{
			if r.limit <= 0{
				break
			}
			n,err := r.s.Read(blks)
			if err == io.EOF && n == 0{
				r.eof = true
				break
			}else if err != nil && err != io.EOF{
				return readb,err
			}
			//fmt.Println("split block send")
			r.splited <- blks[0]
			//fmt.Println("split block send end")
			hash := blks[0].Hash()
			r.rBytesRead += int64(len(blks[0].Data()))
			r.cur = hash[:]
			r.limit --
			r.readblks ++
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
	r.bytesRead += int64(readb)
	return readb,nil
}
//CntReader Read part of the data of content
//
type CntReader struct {
        idxr *IdxReader
        blks chan bblock
        cur []byte

	teeWriter io.Writer	//when read the origin data write it to teeWriter

	bytesRead int64		//bytes readed
}

type bblock struct{
	data []byte
        hash [HashSize]byte
        needUp bool
}

func NewCntReader(idxr *IdxReader,blks chan bblock)*CntReader{
        return &CntReader{
                idxr:idxr,
                blks:blks,
		bytesRead:0,
        }
}

func putUvarint(n uint64) []byte{
  buf := make([]byte, binary.MaxVarintLen64)
  len := binary.PutUvarint(buf,n)
  return buf[:len]
}

func readHelper(r io.Reader,b []byte)(int,error){
  readed := 0
  for{
    n,err := r.Read(b[readed:])
    if err == io.EOF && n == 0{
      break
    } else if err != nil && err != io.EOF{
      return readed,err
    }
    readed += n
    if readed == len(b){
      break
    }
  }
  if readed == 0{
    return 0,io.EOF
  }
  return readed,nil
}

func (r *CntReader)Read(b []byte)(int,error){
        readed := 0
        if len(b) == 0 {
          return 0,nil
        }
        for{
          if len(r.cur) == 0{
		//fmt.Println("bblock read ")
            blk,ok := <- r.blks
		//fmt.Println("bblock read end")
            if ok{
              r.cur = blk.hash[:]

		if r.teeWriter != nil{
			_,err := r.teeWriter.Write(blk.data)
			if err != nil{
				return readed,err
			}
		}

		//fmt.Printf("send hash [%x]\n",r.cur)
              if blk.needUp{
                r.cur = append(r.cur,putUvarint(uint64(len(blk.data)))...)
                r.cur = append(r.cur,blk.data...)
              }else{
                r.cur = append(r.cur,putUvarint(uint64(0))...)
              }
            }else{
              break
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
	  //fmt.Println("cnt read EOF")
          return 0,io.EOF
        }
	r.bytesRead += int64(readed)
	//fmt.Println("Cnt read End")
        return readed,nil
}


type Client struct{
	splited chan Block
	blks	chan bblock
	idxr	*IdxReader
	cntr	*CntReader
	
	bclosed bool
	url string
	getblks int	//index response blocks count
	
	err error	//error stat
	mu sync.Mutex   // Mutex for error state
}

const bufsize = 1024

func NewClient(r io.Reader,url string)*Client{
	splited := make(chan Block,bufsize)
	blks := make(chan bblock,bufsize)
	idxr := NewIdxReader(r,splited,bufsize)
	cntr := NewCntReader(idxr,blks)
	return &Client{
		splited:splited,
		blks:blks,
		idxr:idxr,
		cntr:cntr,
		bclosed:false,
		url:url,
		getblks:0,
	}
}
func NewClientTee(r io.Reader,url string,w io.Writer)*Client{
	cli := NewClient(r,url)
	cli.cntr.teeWriter = w
	return cli
}
func (c *Client)Start(){
	go func(){
		for{
			if c.idxr.eof{
				close(c.splited)
				break
			}
			err := c.idxUpload()
			if err != nil{
				c.setErr(err)
				break
			}
			if c.idxr.eof && c.getblks == c.idxr.readblks{
				close(c.blks)
			}
		}
	}()
}
func (c *Client)setErr(err error){
	if err == nil{
		return
	}
	c.mu.Lock()
	c.err = err
	c.mu.Unlock()
}
func (c *Client)idxUpload()error{
	c.idxr.limit = bufsize
	req,err := http.NewRequest("POST",c.url,c.idxr)
        if err != nil{
                return err
        }
        client := &http.Client{}
        res,err := client.Do(req)
        if err != nil {
                return err
        }
        if res.StatusCode != http.StatusOK {
                err = fmt.Errorf("bad status:%s",res.Status)
                return err
        }
	buf := make([]byte,bufsize,bufsize)
	for{
		//fmt.Println("read response")
		n,err := res.Body.Read(buf)
		//fmt.Println("read response end")
		if err == io.EOF && n == 0{
			break
		}
       		if err != nil && err != io.EOF{
			return err
        	}
		fmt.Println("received ",string(buf[:n]))
		for i := 0;i<n;i++{
			//fmt.Println("split block read")
			spb,ok := <-c.splited
			//fmt.Println("split block read end")
			if !ok{
				return fmt.Errorf("splited chan closed unexpectly")
			}
			needUp := true
			if buf[i] == '1'{
				needUp = false
			}else if buf[i] != '0'{
				return fmt.Errorf("receive format wrong")
			}
			bblk := bblock{
				data:spb.Data(),
				hash:spb.Hash(),
				needUp:needUp,
			}
			//fmt.Println("bblock send ")
			c.blks <- bblk
			//fmt.Println("bblock send end")
			c.getblks ++
		}
	}
	return nil
}
func (c *Client)Read(b []byte)(int,error){
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	if err != nil{
		panic(err)
		return 0,err
	}

	n,err := c.cntr.Read(b)

        c.mu.Lock()
        err2 := c.err
        c.mu.Unlock()
        if err2 != nil{
                return 0,err2
        }
	return n,err
}

func (c *Client)IdxBytesRead()int64{
	return c.idxr.bytesRead
}
func (c *Client)CntBytesRead()int64{
	return c.cntr.bytesRead
}
func (c *Client)ReaderBytesRead()int64{
	return c.idxr.rBytesRead
}
