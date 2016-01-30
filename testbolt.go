package main

import(
	"github.com/boltdb/bolt"
	"fmt"
	"log"
)


const N  = 100000


func testwrite(db *bolt.DB){
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		for i := 0;i<N;i++{
			err := b.Put([]byte(fmt.Sprintf("answer%d",i)), []byte(fmt.Sprintf("%d",i)))
			if err != nil{
				return err
			}
		}
		return nil
	})
}

func testread(db *bolt.DB){
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		for i := 0;i<N;i++{
			v := b.Get([]byte(fmt.Sprintf("answer%d",i)))
			fmt.Printf("The answer is: %s\n", v)
		}
		return nil
	})
}


func main(){
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	db.Update(func(tx *bolt.Tx) error {
	    _, err := tx.CreateBucket([]byte("MyBucket"))
	    if err != nil {
	        return fmt.Errorf("create bucket: %s", err)
	    }
	    return nil
	})
	testwrite(db)
	testread(db)
	defer db.Close()	
}

