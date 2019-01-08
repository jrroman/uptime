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
    Filename    string
    Delay       time.Duration
    Nworkers    int
}

func (a *App) Initialize() {
    flag.StringVar(&a.Filename, "f", "", "Path to the file with sitenames and urls")
    flag.DurationVar(&a.Delay, "d", time.Second * 5, "Delay between requests")
    flag.IntVar(&a.Nworkers, "n", 2, "number of worker routines")
    flag.Parse()
}

func (a *App) GetSiteList() {
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

func (a *App) Run() {
    RequestDispatch(a.Nworkers)
    ResponseDispatch(a.Nworkers)

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        os.Exit(1)
    }()

    for {
        log.Info("scanning sites")
        a.GetSiteList()
        time.Sleep(a.Delay)
    }
}
