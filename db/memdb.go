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

type MemDB struct {
    name    string
    mutex   sync.Mutex        
    db      map[string][]byte
}

func (ndb *MemDB) read (key string) ([]byte, error) {
    var val []byte 

    ndb.mutex.Lock()
    val, ok := ndb.db[key]
    ndb.mutex.Unlock()

    if ok {
        return val, nil
    } else {
        return val, errors.New("Key does not exists in the db")
    }
}

func (ndb *MemDB) write (key string, val []byte) (error) {
    ndb.mutex.Lock()
    ndb.db[key] = val
    ndb.mutex.Unlock()
    
    return nil
}

func (ndb MemDB) GetMmode () bool {
    var mmode bool

    val, err := ndb.read(mmodeKey)
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

func (ndb MemDB) SetMmode (mmode string) {
    var val []byte

    if mmode == mmodeNormal || mmode == mmodeMaintenance {
        val = []byte(mmode)
    } else {
        val = []byte(unknownVal)
    } 

    ndb.write(mmodeKey, val)
}

func (ndb MemDB) InitDB (name string) error {
    ndb.name = name
    ndb.db = make(map[string][]byte)

    return nil
}
