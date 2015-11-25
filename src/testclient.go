package main 

import ( 
  "fmt" 
	"io/ioutil"
  "net/http" 
  "os" 
  "io"
	"time"
  ) 

type  slowReader struct{
	waittime time.Duration
	count int
}

func (r *slowReader)Read(b []byte)(int,error){
	if r.count == 0 {return 0,io.EOF}
	for i,_ := range b{
		b[i] = 'h'
	}
	time.Sleep(r.waittime)
	fmt.Printf("send %d byte \n",len(b))
	return len(b),nil
}

func postHash(url string,file string){
	slowr := &slowReader{
		waittime:3*time.Second,
		count:10,
	}
	req,err := http.NewRequest("POST",url,slowr)
	if err != nil{
		panic(err)
	}
	client := &http.Client{}
	res,err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode != http.StatusOK {
                err = fmt.Errorf("bad status:%s",res.Status)
		return
        }
	body,err := ioutil.ReadAll(res.Body)
	if err != nil{
		panic(err)
	}
	fmt.Println(string(body))
}

func main(){
	//postFile("test","http://10.10.19.104:8080/upload")
	for i,name := range os.Args{
		if i>0{
			postHash("http://10.10.19.104:8080/hash",name)
		}	
	}
}
