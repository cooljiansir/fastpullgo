package main 

import ( 
  "fmt" 
	"io/ioutil"
  "net/http" 
  "mime/multipart" 
  "bytes" 
  "os" 
  "io"
	"fastpull"
  ) 


func postFile(filename string, target_url string) (*http.Response, error) { 
  body_buf := bytes.NewBufferString("") 
  body_writer := multipart.NewWriter(body_buf) 

  // use the body_writer to write the Part headers to the buffer 
  _, err := body_writer.CreateFormFile("file", filename) 
  if err != nil { 
    fmt.Println("error writing to buffer") 
    return nil, err 
  } 

  // the file data will be the second part of the body 
  fh, err := os.Open(filename) 
  if err != nil { 
    fmt.Println("error opening file") 
    return nil, err 
  } 
  defer fh.Close()
  // need to know the boundary to properly close the part myself. 
  boundary := body_writer.Boundary()
  close_string := fmt.Sprintf("\r\n--%s--\r\n", boundary)
  close_buf := bytes.NewBufferString(close_string)
  // use multi-reader to defer the reading of the file data until writing to the socket buffer. 
  request_reader := io.MultiReader(body_buf, fh, close_buf) 
  /*fi, err := fh.Stat() 
  if err != nil { 
    fmt.Printf("Error Stating file: %s", filename) 
    return nil, err 
  }*/ 
  req, err := http.NewRequest("POST", target_url, request_reader) 
  if err != nil { 
    return nil, err 
  } 

  // Set headers for multipart, and Content Length 
  req.Header.Add("Content-Type", "multipart/form-data; boundary=" + boundary) 
  //req.ContentLength = fi.Size()+int64(body_buf.Len())+int64(close_buf.Len()) 

  return http.DefaultClient.Do(req) 
}

func postHash(url string,file string,isFixed bool){
	mmap := make(fastpull.HashBlockMap)
	fmt.Println("deal ",file)
	if isFixed{
		fastpull.MapFile(mmap,file)
	}else{
		fastpull.MapFile2(mmap,file)
	}
	fmt.Println("finish ",file)
	b := bytes.Buffer{}
	for h,_ := range mmap{
		b.Write(h[:])
		fmt.Printf("[%x] \n",h)
	}
	req,err := http.NewRequest("POST",url,&b)
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
	isFixed := true
	for _,arg := range os.Args{
		if arg == "-d"{
			isFixed = false
			break
		}
	}
	for i,name := range os.Args{
		if i>0{
			if name !="-d" {
				postHash("http://10.10.19.104:8080/hash",name,isFixed)
			}
		}	
	}
}
