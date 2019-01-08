package main

import (
    "bufio"
    "flag"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    log "github.com/sirupsen/logrus"
)

var (
    filename    string
    delay       time.Duration
    nworkers    int
)

func init() {
    flag.StringVar(&filename, "f", "", "Path to the file with sitenames and urls")
    flag.DurationVar(&delay, "d", time.Second * 5, "Delay between requests")
    flag.IntVar(&nworkers, "n", 2, "number of worker routines")
    flag.Parse()
}

func GetSitesToScan() {
    f, err := os.Open(filename)
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
       }

       WorkQueue <- work
       log.Info("work request queued")
   }

   if err := scanner.Err(); err != nil {
       log.Info("failed reading file: ", err)
       return
   }
}

func main() {
   RequestDispatch(nworkers)
   ResponseDispatch(nworkers)

   c := make(chan os.Signal)
   signal.Notify(c, os.Interrupt, syscall.SIGTERM)
   go func() {
       <-c
       os.Exit(1)
   }()

   for {
       log.Info("scanning sites")
       GetSitesToScan()
       time.Sleep(delay)
   }
}

