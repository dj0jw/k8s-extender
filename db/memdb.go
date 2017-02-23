package netdb

import (
    "sync"
    "errors"
)

const (
    mmodeKey            = "mmode"
    mmodeNormal         = "Normal"
    mmodeMaintenance    = "Maintenance"
    unknownVal          = "Unknown"
)

var SchedDB *MemDB

type MemDB struct {
    Name    string
    mutex   sync.Mutex        
    db      map[string][]byte
}

func (mdb *MemDB) read (key string) ([]byte, error) {
    var val []byte 

    mdb.mutex.Lock()
    val, ok := mdb.db[key]
    mdb.mutex.Unlock()

    if ok {
        return val, nil
    } else {
        return val, errors.New("Key does not exists in the db")
    }
}

func (mdb *MemDB) write (key string, val []byte) (error) {
    //fmt.Printf("write memdb=%p db=%p\n", mdb, mdb.db)
    mdb.mutex.Lock()
    mdb.db[key] = val
    mdb.mutex.Unlock()
    
    return nil
}

func (mdb *MemDB) GetMmode () bool {
    var mmode bool

    val, err := mdb.read(mmodeKey)
    if err != nil {
        mmode = false
    } else {
        if string(val) == mmodeMaintenance {
            mmode = true
        } else {
            mmode = false  
        }
    }

    return mmode
}

func (mdb *MemDB) SetMmode (mmode string) {
    var val []byte

    if mmode == mmodeNormal || mmode == mmodeMaintenance {
        val = []byte(mmode)
    } else {
        val = []byte(unknownVal)
    } 

    //fmt.Printf("memdb set memdb=%p db=%p\n", mdb, mdb.db)
    mdb.write(mmodeKey, val)
}

func (mdb *MemDB) InitDB (name string) error {
    mdb.Name = name
    mdb.db = make(map[string][]byte, 10)
    //fmt.Printf("init mdb=%p db=%p\n", mdb, mdb.db)

    return nil
}
