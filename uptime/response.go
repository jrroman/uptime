package main

import (
    log "github.com/sirupsen/logrus"
)

var ResponseQueue = make(chan SiteResponse, 50)

type SiteResponse struct {
    URL     string
    Status  int
}

type ResponseWorker struct {
    ID          int
    Work        chan SiteResponse
    WorkerQueue chan chan SiteResponse
    QuitChan    chan bool
}

func NewResponseWorker(id int, workerQueue chan chan SiteResponse) ResponseWorker {
    worker := ResponseWorker{
        ID: id,
        Work: make(chan SiteResponse),
        WorkerQueue: workerQueue,
        QuitChan: make(chan bool),
    }

    return worker
}

func (w *ResponseWorker) Start() {
    go func() {
        for {
            w.WorkerQueue <- w.Work

            select {
            case work := <-w.Work:
                log.Info("checking response")
                work.CheckSiteResponse()
            case <-w.QuitChan:
                log.Info("worker stopping")
                return
            }
        }
    }()
}

func (w *ResponseWorker) Stop() {
    go func() {
        w.QuitChan <- true
    }()
}

func ResponseDispatch(nworkers int) {
    WorkerQueue := make(chan chan SiteResponse, nworkers)

    for i := 0; i < nworkers; i++ {
        log.Info("starting response worker ", i)
        worker := NewResponseWorker(i, WorkerQueue)
        worker.Start()
    }

    go func() {
        for {
            select {
            case work := <-ResponseQueue:
                log.Info("recieved response work")
                go func() {
                    worker := <-WorkerQueue
                    log.Info("dispatching response work request")
                    worker <- work
                }()
            }
        }
    }()
}

func (sr *SiteResponse) CheckSiteResponse() {
    if sr.Status > 399 {
        log.WithFields(log.Fields{
            "URL": sr.URL,
            "Status": sr.Status,
        }).Warn("BAD REQUEST SEND EMAIL")
    } else {
        log.WithFields(log.Fields{
            "URL": sr.URL,
            "Status": sr.Status,
        }).Info("URL OK")
    }
}
