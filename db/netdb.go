package netdb

const (
    MEMORY  = iota
    FILE    = iota
    ELASTIC = iota
)

type NetDB interface {
    InitDB(string) error
    GetMmode () bool
    SetMmode (string)
}

