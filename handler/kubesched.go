package handler

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "github.com/nxoscoder/k8s-extender/utils"
    "github.com/nxoscoder/k8s-extender/db"
    "github.com/nxoscoder/k8s-extender/scheduler"
    schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api/v1"
)

type NetSched struct {
    name        string
    extender    *algorithm.Extender
    db          netdb.NetDB
}

func getSchedulerArgs (rw http.ResponseWriter, req *http.Request) (schedulerapi.ExtenderArgs, error) {
    var args schedulerapi.ExtenderArgs

    decoder := json.NewDecoder(req.Body)
    defer req.Body.Close()

    if err := decoder.Decode(&args); err != nil {
        http.Error(rw, "Decode error", http.StatusBadRequest)
        return args, err
    }

    return args, nil
}

func (ns *NetSched) networkSchedulerFilterHdl (rw http.ResponseWriter, req *http.Request) {
    fmt.Printf("from %s in %s\n", req.URL.Path, utils.FuncName())

    var args schedulerapi.ExtenderArgs

    args, err := getSchedulerArgs(rw, req)
    if err != nil {
        return
    }

    resp := &schedulerapi.ExtenderFilterResult{}
    nodes, failedNodes, err := ns.extender.Filter(&args.Pod, &args.Nodes)
    for _, node := range nodes.Items {
        fmt.Printf("%s\n", node.Name)
    }
    if err != nil {
        resp.Error = err.Error()
    } else {
        resp.Nodes = *nodes
        resp.FailedNodes = failedNodes
    }

    encoder := json.NewEncoder(rw)
    if err := encoder.Encode(resp); err != nil {
        log.Fatalf("Failed to encode %+v", resp)
    }

    return
}

func (ns *NetSched) networkSchedulerPrioritizeHdl (rw http.ResponseWriter, req *http.Request) {
    fmt.Printf("from %s in %s\n", req.URL.Path, utils.FuncName())

    var args schedulerapi.ExtenderArgs

    args, err := getSchedulerArgs(rw, req)
    if err != nil {
        return
    }

    priorities, _ := ns.extender.Prioritize(&args.Pod, &args.Nodes)

    encoder := json.NewEncoder(rw)
    if err := encoder.Encode(priorities); err != nil {
        log.Fatalf("Failed to encode %+v", priorities)
    }

    return
}

func GetKubeSchedulerMuxHdl (ndb netdb.NetDB) (http.Handler) {
    netextender := algorithm.GetExtender()

    netsched := &NetSched {
        name:       "netsched",
        extender:   netextender,
        db:         ndb,
    }

    mux := http.NewServeMux()
    mux.HandleFunc(FilterUrl, netsched.networkSchedulerFilterHdl)
    mux.HandleFunc(PrioritizeUrl, netsched.networkSchedulerPrioritizeHdl)

    return mux
}
