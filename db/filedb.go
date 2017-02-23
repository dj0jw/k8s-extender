package netdb 

import (
    "log"
    "fmt"
    "github.com/boltdb/bolt"
    //"github.com/nxoscoder/k8s-extender/utils"
)

const (
    netResourceDb       = "netresource.db"
    netResourceBucket   = "NetworkResource"
    maxValueLen         = 256
)

type FileDB struct {
    Name    string
    hdl     *bolt.DB
}

func (fdb *FileDB) open() (error) {
    fdb.hdl = nil

    db, err := bolt.Open(netResourceDb, 0600, nil)
    if err != nil {
        log.Fatal(err)
        return err
    }

    fdb.hdl = db

    return fdb.hdl.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucket([]byte(netResourceBucket))
        if err != nil {
            fmt.Errorf("create bucket: %s", err)
            return err
        }
        return nil
    })
}

func (fdb *FileDB) close () (error) {
    fdb.hdl.Close()
    return nil
}

func (fdb *FileDB) write (key []byte, val []byte) (error) {
    return fdb.hdl.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(netResourceBucket))
        err := b.Put([]byte(key), []byte(val))
        //fmt.Printf("put mmode=%s\n", val)
        return err
    })
}

func (fdb *FileDB) read (key string) ([]byte, error) {
    var v []byte

    fdb.hdl.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(netResourceBucket))
        v = b.Get([]byte(key))
        //fmt.Printf("get mmode=%s\n", v)
        return nil 
    })

    return v, nil
}


func (fdb *FileDB) GetMmode () (bool) {
    val, err := fdb.read(mmodeKey)
    if err != nil {
        return false
    } else {
        return getMmodeBool(val)
    }
}

func (fdb *FileDB) SetMmode (mmode string) {
    fdb.write([]byte(mmodeKey), getMmodeVal(mmode))
}

func (fdb *FileDB) OpenDB(name string) (error) {
    fdb.Name = name    
    return fdb.open()
}

func (fdb *FileDB) CloseDB() (error) {
    return fdb.close()
}
