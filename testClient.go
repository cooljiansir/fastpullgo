package main


import(
        "os"
	"fmt"
	"net/http"
        "github.com/cooljiansir/fastpush/client"
)


func main(){
        if len(os.Args) <2 {
                fmt.Println("format test infile ")
                return
        }       
        filestr := os.Args[1]
        file,err := os.Open(filestr)
        if err != nil{
                panic(err)
        }
        clt := client.NewClient(file,"http://127.0.0.1:8080/hash")
        clt.Start()
	if err != nil{
		panic(err)
	}
	req,err := http.NewRequest("POST","http://127.0.0.1:8080/file",clt)
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
                panic(err)
        }
}
