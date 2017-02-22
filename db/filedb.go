package netdb 

import (
    "log"
    "fmt"
    "github.com/boltdb/bolt"
)

const (
    netResourceDb       = "netresource.db"
    netResourceBucket   = "NetworkResource"
)

type FileDB struct {
    hdl *bolt.DB
}

func (fdb *FileDB) Open() (error) {
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

func (fdb *FileDB) Close () (error) {
    fdb.hdl.Close()
    return nil
}

func (fdb *FileDB) Write (key []byte, val []byte) (error) {
    return fdb.hdl.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(netResourceBucket))
        err := b.Put([]byte(key), []byte(val))
        return err
    })
}

func (fdb *FileDB) Read (key string) ([]byte, error) {
    val := make([]byte, 128)

    fdb.hdl.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(netResourceBucket))
        v := b.Get([]byte(key))
        fmt.Printf("mmode=%s\n", v)
        copy(val, v)
        return nil 
    })

    return val, nil
}

