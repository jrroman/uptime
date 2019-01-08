package main

import (
    "fmt"
    "net"
    "net/http"
    "net/url"
    "time"

    log "github.com/sirupsen/logrus"
)

var WorkQueue = make(chan WorkRequest, 100)

type WorkRequest struct {
    Email   string
    Name    string
    URL     string
}

type RequestWorker struct {
    ID          int
    Work        chan WorkRequest
    WorkerQueue chan chan WorkRequest
    QuitChan    chan bool
}

func NewRequestWorker(id int, workerQueue chan chan WorkRequest) RequestWorker {
    worker := RequestWorker{
        ID: id,
        Work: make(chan WorkRequest),
        WorkerQueue: workerQueue,
        QuitChan: make(chan bool),
    }

    return worker
}

func (w *RequestWorker) Start() {
    go func() {
        for {
            w.WorkerQueue <- w.Work

            select {
            case work := <-w.Work:
                fmt.Printf("Worker: %d, name: %s, url: %s\n", w.ID, work.Name, work.URL)
                work.MakeRequest()
            case <-w.QuitChan:
                fmt.Printf("Worker: %d stopping\n", w.ID)
                return
            }
        }
    }()
}

func (w *RequestWorker) Stop() {
    go func() {
        w.QuitChan <- true
    }()
}

func RequestDispatch(nworkers int) {
    WorkerQueue := make(chan chan WorkRequest, nworkers)

    for i := 0; i < nworkers; i++ {
        log.Info("starting worker ", i)
        worker := NewRequestWorker(i, WorkerQueue)
        worker.Start()
    }

    go func() {
        for {
            select {
            case work := <-WorkQueue:
                log.Info("Recieved work request")
                go func() {
                    worker := <-WorkerQueue

                    log.Info("Dispactching work request")
                    worker <- work
                }()
            }
        }
    }()
}

func (wr *WorkRequest) validateURL() string {
    u, err := url.Parse(wr.URL)
    if err != nil {
        log.Fatal("cannot parse URL")
    }

    if u.Scheme == "" {
        u.Scheme = "http"
    }

    return u.String()
}

func (wr *WorkRequest) MakeRequest() {
    validatedURL := wr.validateURL()

    tr := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: time.Second * 5,
        }).DialContext,
    }
    client := &http.Client{
        Timeout: time.Second * 10,
        Transport: tr,
    }

    log.Info("Making request to: ", validatedURL)
    resp, err := client.Head(validatedURL)
    if err != nil {
        log.Warn("request error: ", err)
        return
    }

    response := SiteResponse{
        Email: wr.Email,
        Status: resp.StatusCode,
        URL: validatedURL,
    }

    ResponseQueue <- response
}
