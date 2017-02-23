package netdb

const (
    _       = iota
    MEMORY
    FILE
    ELASTIC
)

const (
    mmodeKey            = "mmode"
    mmodeNormal         = "Normal"
    mmodeMaintenance    = "Maintenance"
    unknownVal          = "Unknown"
)

var SchedDB NetDB

type NetDB interface {
    OpenDB (string) (error)
    CloseDB () (error)
    GetMmode () (bool)
    SetMmode (string)
}

func getMmodeVal (mmode string) ([]byte) {
    var val []byte

    if mmode == mmodeNormal || mmode == mmodeMaintenance {
        val = []byte(mmode)
    } else {
        val = []byte(unknownVal)
    } 

    return val
}

func getMmodeBool (val []byte) (bool) {
    var mmode bool

    if string(val) == mmodeMaintenance {
        mmode = true
    } else {
        mmode = false  
    }

    return mmode
}
