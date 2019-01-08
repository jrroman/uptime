package main

import (
    "bufio"
    "flag"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/sirupsen/logrus"
)

var log = logrus.New()

type App struct {
    Delay       time.Duration
    Filename    string
    Nworkers    int
    Wait        time.Duration
}

func (a *App) Initialize() {
    flag.DurationVar(&a.Delay, "d", time.Second * 5, "Delay between requests")
    flag.StringVar(&a.Filename, "f", "", "Path to the file with sitenames and urls")
    flag.IntVar(&a.Nworkers, "n", 2, "number of worker routines")
    flag.DurationVar(&a.Wait, "w", time.Second * 5, "Delay between requests")
    flag.Parse()
}

func (a *App) PopulateSiteList() {
    f, err := os.Open(a.Filename)
    if err != nil {
       log.Info("failed opening file: ", err)
       return
   }
   defer f.Close()

   scanner := bufio.NewScanner(f)
   for scanner.Scan() {
       data := strings.Split(scanner.Text(), " ")
       work := WorkRequest{
           Name: data[0],
           URL: data[1],
           Email: data[2],
       }

       WorkQueue <- work
       log.Info("work request queued")
   }

   if err := scanner.Err(); err != nil {
       log.Info("failed reading file: ", err)
       return
   }
}

func (a *App) Dispatcher() {
    RequestWorkerQueue := make(chan chan WorkRequest, a.Nworkers)
    ResponseWorkerQueue := make(chan chan SiteResponse, a.Nworkers)

    for i := 0; i < a.Nworkers; i++ {
        log.Info("Starting request worker ", i)
        requestWorker := NewRequestWorker(i, RequestWorkerQueue)
        requestWorker.Start()

        log.Info("Starting response worker ", i)
        responseWorker := NewResponseWorker(i, ResponseWorkerQueue)
        responseWorker.Start()
    }

    go func() {
        for {
            select {
            case work := <-WorkQueue:
                log.Info("Recieved work request")
                go func() {
                    worker := <-RequestWorkerQueue
                    log.Info("Dispatching request worker")
                    worker <- work
                }()
            case work := <-ResponseQueue:
                log.Info("Recieved response work")
                go func() {
                    worker := <-ResponseWorkerQueue
                    log.Info("Dispatching response worker")
                    worker <- work
                }()
            }
        }
    }()
}

func (a *App) Run() {
    a.Dispatcher()

    done := make(chan bool)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-quit
        log.Info("Uptime bot is shutting down")
        time.Sleep(a.Wait)
        close(done)
    }()

    go func() {
        for {
            log.Info("Scanning sites")
            a.PopulateSiteList()
            time.Sleep(a.Delay)
        }
    }()

    <-done
    log.Info("server stopped")
}
