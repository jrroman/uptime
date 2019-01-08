package main

import (
    "net"
    "net/http"
    "net/url"
    "time"

    "github.com/sirupsen/logrus"
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
                log.WithFields(logrus.Fields{
                    "WorkerId": w.ID,
                    "Name": work.Name,
                    "URL": work.URL,
                }).Info("request made")
                work.MakeRequest()
            case <-w.QuitChan:
                log.WithFields(logrus.Fields{
                    "WorkerId": w.ID,
                }).Info("worker stopping")
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
    req, err := http.NewRequest("HEAD", validatedURL, nil)
    if err != nil {
        log.Warn("request error: ", err)
        return
    }
    req.Header.Set("User-Agent", "HearstLab_Uptime_bot/1.0; http://www.hearstlab.com")

    resp, err := client.Do(req)
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
