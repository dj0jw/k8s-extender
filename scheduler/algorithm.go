package algorithm

import (
    "fmt"
    //"github.com/nxoscoder/k8s-extender/db"
    "k8s.io/kubernetes/pkg/api/v1"
    schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api/v1"
)

type FitPredicate func(pod *v1.Pod, node *v1.Node, e *Extender) (bool, error)
type PriorityFunc func(pod *v1.Pod, nodes *v1.NodeList, e *Extender) (*schedulerapi.HostPriorityList, error)

type PriorityConfig struct {
    function PriorityFunc
    weight   int
}

type Extender struct {
    name            string
    predicates      []FitPredicate
    prioritizers    []PriorityConfig
}

func GetExtender () (*Extender) {
   return &Extender{
        name:         "netextender",
        predicates:   []FitPredicate{NetworkPredicate},
        prioritizers: []PriorityConfig{{NetworkPrioritizer, 1}},
    }
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

        PriorityFunc := prioritizer.function
        prioritizedList, err := PriorityFunc(pod, nodes, e)
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
func NetworkPredicate (pod *v1.Pod, node *v1.Node, e *Extender) (bool, error) {
    // if node.Name == "n9kinfra-mr-master01" || node.Name == "n9kinfra-mr-node01" || node.Name == "n9kinfra-mr-node02" {
    if node.Name == "n9kinfra-mr-master01" || node.Name == "n9kinfra-mr-node01" {
        return true, nil
    }

    return false, nil
}

func NetworkPrioritizer (pod *v1.Pod, nodes *v1.NodeList, e *Extender) (*schedulerapi.HostPriorityList, error) {
    result := schedulerapi.HostPriorityList{}

    for _, node := range nodes.Items {
        score := 1
        // Since Kubernetes typically avoids scheduling on the master we will show that when 
        // the network-element is busted, we are increasing the score to get to the master. 
        network_elem := false   // true when in maintenance mode
        fmt.Printf("node=%s\n", node.Name)
        //fmt.Printf("mmode=%t\n", e.db.getMmode())
        if node.Name == "n9kinfra-mr-master01" && network_elem {
            score = 10
        }
        result = append(result, schedulerapi.HostPriority{node.Name, score})
    }

    return &result, nil
}
