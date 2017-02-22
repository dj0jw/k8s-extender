package handler

const (
    SchedulerDom    = "my_network_scheduler"
    HekaDom         = "hekain"
    RootUrl         = "/"
    SchedulerUrl    = RootUrl + SchedulerDom
    FilterUrl       = SchedulerUrl + "/filter"
    PrioritizeUrl   = SchedulerUrl + "/prioritize"
    HekaUrl         = RootUrl + HekaDom
    MmodeUrl        = HekaUrl + "/mmode"
)
