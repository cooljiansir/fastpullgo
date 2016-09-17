package fingerdb

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/binary"
	"bytes"
	"io"
	"github.com/cooljiansir/fastpush/spliter"
	"os"
	"fmt"
)

/*

get(fingerprint)

cache -> writebuffer -> db(read the whole group metadata into cache)

write(fingerprint,metadata)
	->writebuffer
        	flush->disk(file,db)

 




*/


type MetaData struct{
	Offset uint64
	Length uint32
	Containerid uint64
}

func byte2string(b []byte)string {
	t := "0123456789abcdef"
	res := ""
	for i := 0; i < len(b); i++ {
		bb := int(b[i])
		res += string(t[bb / 16])
		res += string(t[bb % 16])
	}
	return res
}

func (m *MetaData)tobyte()([]byte,error){
	buf := new(bytes.Buffer)
	err := binary.Write(buf,binary.BigEndian,m)
	if err != nil{return []byte{},err}

	return buf.Bytes(),nil
}

//read out a meta data
func readMeta(r io.Reader)(MetaData,error){
	var meta MetaData
	err := binary.Read(r,binary.BigEndian,&meta)
	if err != nil{return MetaData{},err}
	return meta,nil
}

//the bucket name bolt.db

const BUCKET = "metadata"
const DBBUCKET = "dbbucket"
const MAXCONTAINER = "maxcontainer"

const dbWriteBufferN = 1024

type FingerDB struct{
	db *bolt.DB
	dbc chan bool		//mutex for db 
	cache map[[spliter.HashSize]byte]MetaData
	dbfile string		//boltdb file location
	basepath string		//file path to store containers
}


func NewFingerDB(dbpath string)(*FingerDB,error){
	dbfile := fmt.Sprintf("%s/finger.db",dbpath)
	if _, err := os.Stat(dbpath); os.IsNotExist(err) {
		if err := os.MkdirAll(dbpath, 0777); err != nil {
			return nil,err
		}
	}
	db, err := bolt.Open(dbfile, 0600, nil)
	if err != nil{
		return nil,err
	}
	dbc := make(chan bool,1)
	dbc <- true
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET))
		if b == nil{
			_, err := tx.CreateBucket([]byte(BUCKET))
    			if err != nil {
        			return err
    			}
		}
		b = tx.Bucket([]byte(DBBUCKET))
		if b == nil{
			_, err := tx.CreateBucket([]byte(DBBUCKET))
    			if err != nil {
        			return err
    			}
		}
		return nil
	}); err != nil {
		<- dbc
    		log.Fatal(err)
	}
	<- dbc
	return &FingerDB{
		db:db,
		cache:make(map[[spliter.HashSize]byte]MetaData),
		dbfile:dbfile,
		basepath:dbpath,
		dbc:dbc,
	},nil
}
func (fdb *FingerDB)BeginTx(){
	//fmt.Println("Enterring...")
	fdb.dbc <- true
	//fmt.Println("Entered...")
}
func (fdb *FingerDB)EndTx(){
	//fmt.Println("Quiting...")
	<- fdb.dbc
	//fmt.Println("Quited")
}

func (fdb *FingerDB)getBlockPath(containerid uint64)string{
	return fmt.Sprintf("%s/%d.blk",fdb.basepath,containerid)
}

func (fdb *FingerDB)getMetaPath(containerid uint64)string{
	return fmt.Sprintf("%s/%d.meta",fdb.basepath,containerid)
}

//db maxContainer++
func (fdb *FingerDB)addMaxContainer()(uint64,error){
	fdb.BeginTx()
	mxc := uint64(0)
	if err := fdb.db.Update(func(tx *bolt.Tx) error {
	    	b := tx.Bucket([]byte(DBBUCKET))
		bmx := b.Get([]byte(MAXCONTAINER))
		if bmx != nil{
			fmt.Sscanf(string(bmx),"%d",&mxc)
		}
		mxc ++
		if err := b.Put([]byte(MAXCONTAINER), []byte(fmt.Sprintf("%d",mxc))); err != nil {
			return err
		}
		return nil
	}); err != nil {
		fdb.EndTx()
		return 0,err
	}
	fdb.EndTx()
	return mxc,nil
}


func (fdb *FingerDB)Find(f [spliter.HashSize]byte)(MetaData,bool){

	//return MetaData{},false

	fdb.BeginTx()
	meta,find := fdb.cache[f]
	if find {
		fdb.EndTx()
		return meta,find
	}
	buf := new(bytes.Buffer)
	if err := fdb.db.View(func(tx *bolt.Tx) error {
		//fmt.Println("Look up-----> ",byte2string(f[:]))
		b := tx.Bucket([]byte(BUCKET))
		value := b.Get(f[:])
		if value == nil {
			find =  false
		}else{
			buf.Write(value)
			find = true
		}
		//fmt.Println("Look up<----- ",byte2string(f[:]))
		return nil
	}); err != nil {
		fdb.EndTx()
		log.Fatal(err)
	}
	fdb.EndTx()
	if !find{
		return MetaData{},find
	}
	meta,err := readMeta(buf)
	if err != nil{
		log.Fatal(err)
		return MetaData{},false
	}
	fmt.Printf("\rFind in disk")
	fdb.cache[f] = meta
	return meta,true
}

const MAXBLOCKSIZE = 1024*1024

type Container struct{
	fingerdb *FingerDB
	containerid uint64
	blkWriteCloser io.WriteCloser
	blkbytes uint64		//the bytes of blks of current file

	//buffer
	dbWriteBuffer chan [][spliter.HashSize]byte
	dbWBufferPre [][spliter.HashSize]byte
	flushed chan bool
}

func (fdb *FingerDB)NewContainer()*Container{
	c := &Container{
		fingerdb:fdb,
		dbWriteBuffer:make(chan [][spliter.HashSize]byte,1),
		dbWBufferPre:[][spliter.HashSize]byte{},
		flushed:make(chan bool),
	}
	c.newFile()
	go c.flushDBBuffer()
	return c
}

func (c *Container)newFile()error{
	mxc,err := c.fingerdb.addMaxContainer()
	if err != nil{
		return err
	}
	blkWriteCloser,err := os.Create(c.fingerdb.getBlockPath(mxc))
	if err != nil{
		return err
	}
	c.containerid = mxc
	c.blkbytes = 0
	c.blkWriteCloser = blkWriteCloser
	return nil
}
func (c *Container)closeFile()error{
	return  c.blkWriteCloser.Close()
}
func (c *Container)Close()error{
	fmt.Println("Closing...")
	close(c.dbWriteBuffer)
	<-c.flushed
	return  c.closeFile()
}
func (c *Container)flushDBBuffer()error{
	for {
		dbbuff,ok := <- c.dbWriteBuffer
		if !ok{
			break
		}
		fmt.Println("Flushing one...")
		c.fingerdb.BeginTx()
		fmt.Println("Updating ...")
        	if err := c.fingerdb.db.Update(func(tx *bolt.Tx) error {
                	b := tx.Bucket([]byte(BUCKET))
			for _,f := range dbbuff{
				m,find := c.fingerdb.cache[f]
				if !find{
					fmt.Printf("%s has been flushed before\n",byte2string(f[:]))
					//return  nil	//flushed
					continue
				}
		        	buf,err := m.tobyte()
			        if err != nil{
        			        return err
        			}
		                if err := b.Put(f[:], buf); err != nil {
	        	                return err
	               	 	}
				//fmt.Println("Put ",byte2string(f[:]))
			}
	                return nil
        	}); err != nil {
			c.fingerdb.EndTx()
                	return err
        	}
		for _,f := range dbbuff{
			delete(c.fingerdb.cache,f)
		}
		fmt.Println("Upadted.")
		c.fingerdb.EndTx()
		fmt.Println("Flushed one.")
	}
	{
		c.fingerdb.BeginTx()
		fmt.Println("Updating ...")
        	if err := c.fingerdb.db.Update(func(tx *bolt.Tx) error {
                	b := tx.Bucket([]byte(BUCKET))
			for _,f := range c.dbWBufferPre{
				m,find := c.fingerdb.cache[f]
				if !find{
					return  nil	//flushed
				}
		        	buf,err := m.tobyte()
			        if err != nil{
        			        return err
        			}
		                if err := b.Put(f[:], buf); err != nil {
	        	                return err
	               	 	}
				//fmt.Println("Put ",byte2string(f[:]))
			}
	                return nil
        	}); err != nil {
			c.fingerdb.EndTx()
                	return err
        	}
		for _,f := range c.dbWBufferPre{
			delete(c.fingerdb.cache,f)
		}
		c.fingerdb.EndTx()
		fmt.Println("Upadted.")
	}
	c.flushed <- true
	fmt.Println("Flush exit.")
	return nil
}
func (c *Container)writeDBBuffer(f [spliter.HashSize]byte,m MetaData)error{
	c.fingerdb.BeginTx()
	c.fingerdb.cache[f] = m
	c.dbWBufferPre = append(c.dbWBufferPre,f)
	if len(c.dbWBufferPre) >= dbWriteBufferN{
		c.dbWriteBuffer <- c.dbWBufferPre
		c.dbWBufferPre = [][spliter.HashSize]byte{}
	}
	c.fingerdb.EndTx()
	return nil
}

func (c *Container)WriteBlock(f [spliter.HashSize]byte,blk []byte)error{
	// Insert data into a bucket.
	m := MetaData{
		Offset:c.blkbytes,
		Length:uint32(len(blk)),
		Containerid:c.containerid,
	}

	n,err := c.blkWriteCloser.Write(blk)

	if err != nil{
		return err
	}

	c.blkbytes += uint64(n)
	
	if c.blkbytes >= MAXBLOCKSIZE {
		fmt.Println("new file")
		err = c.closeFile()
		if err != nil{
			return err
		}
		err = c.newFile()
		if err != nil{
			return err
		}
	}

	return c.writeDBBuffer(f,m)
}


type BlockReader struct {
	metadata MetaData
	treaded uint32
	readcloser io.ReadCloser
}

func (fdb *FingerDB)NewBlockReader(metadata MetaData)(*BlockReader,error){
	readcloser,err := os.Open(fdb.getBlockPath(metadata.Containerid))
	if err != nil{
		return nil,err
	}
	offset := int64(metadata.Offset)
	_,err = readcloser.Seek(offset,0)
	if err != nil{
		return nil,err
	}
	return &BlockReader{
		metadata:metadata,
		treaded:0,
		readcloser:readcloser,
	},nil
}


func (r *BlockReader)Read(b []byte)(int,error){
	return r.readcloser.Read(b)
}

func (r *BlockReader)Close()error{
	return r.readcloser.Close()
}

