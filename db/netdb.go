package netdb

const (
    _       = iota
    MEMORY
    FILE
    ELASTIC
)

type NetDB interface {
    InitDB(string) error
    GetMmode () bool
    SetMmode (string)
}
