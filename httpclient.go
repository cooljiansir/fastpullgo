package main 

import ( 
  "fmt" 
	"io/ioutil"
  "net/http" 
  "os" 
	"github.com/cooljiansir/fastpush/client"
  ) 



func postHash(url string,file string){
	f,err := os.Open(file)
	if err != nil{
		panic(err)
	}
	defer f.Close()
	b := client.NewIdxReader(f)
	req,err := http.NewRequest("POST",url,b)
	if err != nil{
		panic(err)
	}
	client := &http.Client{}
	fmt.Println("debug ")
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
	if len(os.Args) < 2 {
		fmt.Println("format: test file")
		return
	}
	file := os.Args[1]
	postHash("http://127.0.0.1:8080/hash",file)
}
