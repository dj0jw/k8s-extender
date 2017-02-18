package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    "runtime"
    "io/ioutil"
    "lib"
    "sync"
    "errors"
    "k8s.io/kubernetes/pkg/api/v1"
    schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api/v1"
    //_ "plugin/pkg/scheduler/algorithmprovider"
    //schedulerapi "plugin/pkg/scheduler/api/v1"
    //"k8s.io/kubernetes/pkg/api/v1"
    //_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"
    //schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api/v1"
    //"k8s.io/kubernetes/pkg/api/resource"
    //"k8s.io/kubernetes/pkg/apimachinery/registered"
    //metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
    //clientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
    //v1core "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/core/v1"
    //"k8s.io/kubernetes/pkg/client/record"
    //"k8s.io/kubernetes/pkg/client/restclient"
    //"k8s.io/kubernetes/pkg/util/wait"
    //"k8s.io/kubernetes/plugin/pkg/scheduler"
    //"k8s.io/kubernetes/plugin/pkg/scheduler/factory"
    //e2e "k8s.io/kubernetes/test/e2e/framework"
    //"k8s.io/kubernetes/test/integration/framework"
)

type fitPredicate func(pod *v1.Pod, node *v1.Node, e *Extender) (bool, error)
type priorityFunc func(pod *v1.Pod, nodes *v1.NodeList, e *Extender) (*schedulerapi.HostPriorityList, error)

type priorityConfig struct {
    function priorityFunc
    weight   int
}

type Extender struct {
    name         string
    predicates   []fitPredicate
    prioritizers []priorityConfig
    db           *networkDB
}

// For now we assume that the Nodes that we are running on have names  
// "n9kinfra-mr-master01", "n9kinfra-mr-node01" and are present inthe NodeList
// We need to obtain the name to network-element translation from the database to work of network-elements
func (e *Extender) Filter (pod *v1.Pod, nodes *v1.NodeList) (*v1.NodeList, schedulerapi.FailedNodesMap, error) {
    filtered := []v1.Node{}
    failedNodesMap := schedulerapi.FailedNodesMap{}

    for _, node := range nodes.Items {
        fits := true
        for _, predicate := range e.predicates {
            fit, err := predicate(pod, &node, e)
            if err != nil {
                return &v1.NodeList{}, schedulerapi.FailedNodesMap{}, err
            }
            if !fit {
                fits = false
                break
            }
        }
        if fits {
            filtered = append(filtered, node)
        } else {
            failedNodesMap[node.Name] = fmt.Sprintf("extender failed: %s", e.name)
        }
    }

    return &v1.NodeList{Items: filtered}, failedNodesMap, nil
}

func (e *Extender) Prioritize (pod *v1.Pod, nodes *v1.NodeList) (*schedulerapi.HostPriorityList, error) {
    result := schedulerapi.HostPriorityList{}
    combinedScores := map[string]int{}

    for _, prioritizer := range e.prioritizers {
        weight := prioritizer.weight
        if weight == 0 {
            continue
        }
        priorityFunc := prioritizer.function
        prioritizedList, err := priorityFunc(pod, nodes, e)
        if err != nil {
            return &schedulerapi.HostPriorityList{}, err
        }
        for _, hostEntry := range *prioritizedList {
            combinedScores[hostEntry.Host] += hostEntry.Score * weight
        }
    }

    for host, score := range combinedScores {
        fmt.Printf("host=%s score=%d\n", host, score)
        result = append(result, schedulerapi.HostPriority{Host: host, Score: score})
    }

    return &result, nil
}

// Initial idea was to allow for the master and node-01 to pass predicates and fail node-02
// However, for now we have only the master and node-01. If needed we can re-compile on-line
// to show the predicates is dropping once, and weights is helping in another test run
func machine_Predicate (pod *v1.Pod, node *v1.Node, e *Extender) (bool, error) {
    // if node.Name == "n9kinfra-mr-master01" || node.Name == "n9kinfra-mr-node01" || node.Name == "n9kinfra-mr-node02" {
    if node.Name == "n9kinfra-mr-master01" || node.Name == "n9kinfra-mr-node01" {
        return true, nil
    }

    return false, nil
}

func machine_Prioritizer (pod *v1.Pod, nodes *v1.NodeList, e *Extender) (*schedulerapi.HostPriorityList, error) {
    result := schedulerapi.HostPriorityList{}

    for _, node := range nodes.Items {
        score := 1
        // Since Kubernetes typically avoids scheduling on the master we will show that when 
        // the network-element is busted, we are increasing the score to get to the master. 
        network_elem := false   // true when in maintenance mode
        fmt.Printf("node=%s\n", node.Name)
        fmt.Printf("mmode=%t\n", e.db.getMmode())
        if node.Name == "n9kinfra-mr-master01" && network_elem {
            score = 10
        }
        result = append(result, schedulerapi.HostPriority{node.Name, score})
    }

    return &result, nil
}

const (
    schedulerDom    = "my_network_scheduler"
    hekaDom         = "hekain"
    rootUrl         = "/"
    schedulerUrl    = rootUrl + schedulerDom
    filterUrl       = schedulerUrl + "/filter"
    prioritizeUrl   = schedulerUrl + "/prioritize"
    hekaUrl         = rootUrl + hekaDom
    mmodeUrl        = hekaUrl + "/mmode"
)

func funcName() string {
    pc, _, _, _ := runtime.Caller(1)
    return runtime.FuncForPC(pc).Name()
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

func (extender *Extender) networkSchedulerFilterHdl (rw http.ResponseWriter, req *http.Request) {
    fmt.Printf("from %s in %s\n", req.URL.Path, funcName())

    var args schedulerapi.ExtenderArgs

    args, err := getSchedulerArgs(rw, req)
    if err != nil {
        return
    }

    resp := &schedulerapi.ExtenderFilterResult{}
    nodes, failedNodes, err := extender.Filter(&args.Pod, &args.Nodes)
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

func (extender *Extender) networkSchedulerPrioritizeHdl (rw http.ResponseWriter, req *http.Request) {
    fmt.Printf("from %s in %s\n", req.URL.Path, funcName())

    var args schedulerapi.ExtenderArgs

    args, err := getSchedulerArgs(rw, req)
    if err != nil {
        return
    }

    priorities, _ := extender.Prioritize(&args.Pod, &args.Nodes)

    encoder := json.NewEncoder(rw)
    if err := encoder.Encode(priorities); err != nil {
        log.Fatalf("Failed to encode %+v", priorities)
    }

    return
}

type SubdomainsMux map[string]http.Handler

func (subdomains SubdomainsMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("host=%s path=%s\n", r.Host, r.URL.Path) 

    var domParts []string
    for _, part := range strings.Split(r.URL.Path, "/") {
        if part != "" {
            domParts = append(domParts, part)
        }
    }

    for i, d := range domParts {
        fmt.Printf("[%d] %s\n", i, d)
    }

    fmt.Printf("domain=%s\n", domParts[len(domParts)-1])

    if mux := subdomains[domParts[0]]; mux != nil {
        r.URL.Path = "/" + strings.Join(domParts, "/")
        fmt.Printf("host=%s path=%s\n", r.Host, r.URL.Path) 
        mux.ServeHTTP(w, r)
    } else {
        http.Error(w, "Not found", 404)
    }

    return
}

func getKubeSchedulerMuxHdl (ndb *networkDB) (http.Handler) {
    extender := &Extender{
        name:         "extender",
        predicates:   []fitPredicate{machine_Predicate},
        prioritizers: []priorityConfig{{machine_Prioritizer, 1}},
        db: ndb,
    }

    mux := http.NewServeMux()
    mux.HandleFunc(filterUrl, extender.networkSchedulerFilterHdl)
    mux.HandleFunc(prioritizeUrl, extender.networkSchedulerPrioritizeHdl)

    return mux
}

type HekaIn struct {
    name    string
    db      *networkDB
}

func (hi *HekaIn) mmodeHdl (rw http.ResponseWriter, req *http.Request) {
    fmt.Printf("from %s in %s\n", req.URL.Path, funcName())

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
            hi.db.setMmode(mmode.(string))
        }
    }

    return
}

func getHekaInMuxHdl (ndb *networkDB) (http.Handler) {
    hekain := &HekaIn{name: "hekaIn", db: ndb}

    mux := http.NewServeMux()
    mux.HandleFunc(mmodeUrl, hekain.mmodeHdl)

    return mux
}

const (
    mmodeKey            = "mmode"
    mmodeNormal         = "Normal"
    mmodeMaintenance    = "Maintenance"
    unknownVal          = "Unknown"
)

type networkDB struct {
    mutex   sync.Mutex        
    db      map[string][]byte
}

func (ndb *networkDB) read (key string) ([]byte, error) {
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

func (ndb *networkDB) write (key string, val []byte) (error) {
    ndb.mutex.Lock()
    ndb.db[key] = val
    ndb.mutex.Unlock()
    
    return nil
}

func (ndb *networkDB) getMmode () bool {
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

func (ndb *networkDB) setMmode (mmode string) {
    var val []byte

    if mmode == mmodeNormal || mmode == mmodeMaintenance {
        val = []byte(mmode)
    } else {
        val = []byte(unknownVal)
    } 

    ndb.write(mmodeKey, val)
}

func main() {
    fmt.Printf("Main function for extender\n")

    var sdb db.SchedDb
    err := sdb.Open()
    if err != nil {
        fmt.Printf("Failed to open db\n")
        return
    }

    ndb := networkDB{db: make(map[string][]byte)}

    subdomains := make(SubdomainsMux)
    subdomains[schedulerDom] = getKubeSchedulerMuxHdl(&ndb) 
    fmt.Printf("Register mux for %s\n", schedulerDom)
    subdomains[hekaDom] = getHekaInMuxHdl(&ndb) 
    fmt.Printf("Register mux for %s\n", hekaDom)

    err = http.ListenAndServe(":50000", subdomains)
    if err != nil {
        fmt.Printf("Failed to listen and serve\n")
    }
}
