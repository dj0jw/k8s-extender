package main

import (
    "fmt"
    "net/http"
    "strings"
    "github.com/nxoscoder/k8s-extender/db"
    "github.com/nxoscoder/k8s-extender/handler"
)

const (
    schedulerDom    = "my_network_scheduler"
    hekaDom         = "hekain"
)

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

func main() {
    fmt.Printf("Main function for extender\n")

    ndb := netdb.MemDB{}
    ndb.InitDB("memdb")

    subdomains := make(SubdomainsMux)
    subdomains[handler.SchedulerDom] = handler.GetKubeSchedulerMuxHdl(ndb) 
    fmt.Printf("Register mux for %s\n", handler.SchedulerDom)
    subdomains[handler.HekaDom] = handler.GetHekaInMuxHdl(ndb) 
    fmt.Printf("Register mux for %s\n", handler.HekaDom)

    err := http.ListenAndServe(":50000", subdomains)
    if err != nil {
        fmt.Printf("Failed to listen and serve\n")
    }
}
