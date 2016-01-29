package fingerdb

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/binary"
	"bytes"
	"io"
	"github.com/cooljiansir/fastpush/spliter"
)

/*

get(fingerprint)

cache -> writebuffer -> db(read the whole group metadata into cache)

write(fingerprint,metadata)
	->writebuffer
        	flush->disk(file,db)

 




*/


type MetaData struct{
	offset uint64
	length uint32
	groupno	uint32
	filename string
}

// []offset uint64
// []length uint32
// []groupno uint32
// []strlen uint16
// []filename string
func (m *MetaData)tobyte()([]byte,error){
	buf := new(bytes.Buffer)
	err := binary.Write(buf,binary.BigEndian,m.offset)
	if err != nil{return []byte{},err}

	err = binary.Write(buf,binary.BigEndian,m.length)
	if err != nil{return []byte{},err}
	
	err = binary.Write(buf,binary.BigEndian,m.groupno)
	if err != nil{return []byte{},err}

	strlen := uint16(len(m.filename))

	err = binary.Write(buf,binary.BigEndian,strlen)
	if err != nil{return []byte{},err}

	buf.WriteString(m.filename)
	return buf.Bytes()
}

func readhelper(r io.Reader,b []byte)(int,error){
	readed := 0
	for{
		n,err := r.Read(b[readed:])
		readed += n
		if err != nil{
			return readed,err
		}
		if readed == len(b){
			return readed,nil
		}
	}
}
//read out a meta data
func readMeta(r io.Reader)(MetaData,error){
	var offset uint64
	var length uint32
	var groupno uint32
	var strlen uint16
	err := binary.Read(r,binary.BigEndian,&offset)
	if err != nil{return MetaData{},err}
	
	err = binary.Read(r,binary.BigEndian,&length)
	if err != nil{return MetaData{},err}

	err = binary.Read(r,binary.BigEndian,&groupno)
	if err != nil{return MetaData{},err}

	err = binary.Read(r,binary.BigEndian,&strlen)
	if err != nil{return MetaData{},err}

	strl := int(strlen)

	buf := make([]byte,strl,strl)
	n,err := readhelper(r,buf)
	if n!= strl{
		return MetaData{},fmt.Errorf("length is not as expected")
	}
	if err != nil{
		return err
	}
	filename := string(buf)
	return MetaData{
		offset:offset,
		length:length,
		groupno:groupno,
		filename:filename,
	}
}

//the bucket name bolt.db

const BUCKET = "metadata"

type FingerDB struct{
	db bolt.DB
	cache map[[spliter.HashSize]byte]MetaData
	dbfile string		//boltdb file location
	basepath string		//file path to store meta data
}


func (fdb *FingerDB)find(f [spliter.HashSize]byte)(MetaData,bool){
	meta,find := fdb.cache[f]
	if find {
		return meta,find
	}

	syn := make(chan bool)
	buf := new(bytes.Buffer)
	if err := fdb.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(BUCKET)).Get(f)
		if value == nil {
			syn <- false
		}else{
			buf.Write(value)
			syn <- true
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	
	find <- syn
	if !find{
		return MetaData{},find
	}
	readMeta(buf)
}
