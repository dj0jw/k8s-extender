package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "io/ioutil"
    "github.com/nxoscoder/k8s-extender/utils"
    "github.com/nxoscoder/k8s-extender/db"
)

type HekaIn struct {
    name    string
}

func (hi *HekaIn) mmodeHdl (rw http.ResponseWriter, req *http.Request) {
    fmt.Printf("from %s in %s\n", req.URL.Path, utils.FuncName())

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        panic(err)
    }

    //fmt.Print(string(body))

    var data map[string]interface{}
    err = json.Unmarshal(body, &data)
    if err != nil {
        panic(err)
    }

    if output, ok := data["output"]; ok {
        if mmode, ok := output.(map[string]interface{})["system_mode"]; ok {
            fmt.Printf("mmode=%s\n", mmode)
            //fmt.Printf("heka scheddb=%p\n", netdb.SchedDB)
            netdb.SchedDB.SetMmode(mmode.(string))
        }
    }

    return
}

func GetHekaInMuxHdl () (http.Handler) {
    hekain := &HekaIn{name: "hekain"}

    mux := http.NewServeMux()
    mux.HandleFunc(MmodeUrl, hekain.mmodeHdl)

    return mux
}

