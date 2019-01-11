package main

import (
	"bufio"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// w "github.com/jrroman/uptime/uptime/worker"
	w "hlab/uptime/worker"

	"github.com/sirupsen/logrus"
)

type App struct {
	Delay    time.Duration
	Done     chan bool
	Filename string
	Logger   *logrus.Logger
	Nworkers int
	Quit     chan os.Signal
	Wait     time.Duration
}

func (a *App) Initialize() {
	a.Done = make(chan bool)
	a.Quit = make(chan os.Signal, 1)
	flag.DurationVar(&a.Delay, "d", time.Second*5, "Delay between requests")
	flag.StringVar(&a.Filename, "f", "", "Path to the file with sitenames and urls")
	flag.IntVar(&a.Nworkers, "n", 2, "number of worker routines")
	flag.DurationVar(&a.Wait, "w", time.Second*5, "Delay between requests")
	flag.Parse()
	a.InitializeLogger()
}

func (a *App) InitializeLogger() {
	a.Logger = logrus.New()
}

func (a *App) PopulateSiteList() {
	f, err := os.Open(a.Filename)
	if err != nil {
		a.Logger.Info("failed opening file: ", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), " ")

		name := data[0]
		URL := w.ValidateURL(data[1])
		email := data[2]

		work := w.WorkRequest{
			Email: email,
			Name:  name,
			Type:  "request",
			URL:   URL,
		}
		w.WorkQueue <- work
		a.Logger.Info("Work request queued")
	}

	if err := scanner.Err(); err != nil {
		a.Logger.Info("failed reading file: ", err)
		return
	}
}

func (a *App) Dispatcher() {
	workerQueue := make(chan chan w.WorkRequest, a.Nworkers)

	for i := 0; i < a.Nworkers; i++ {
		a.Logger.Info("Starting worker #", i+1)
		worker := w.NewWorker(i+1, workerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-w.WorkQueue:
				a.Logger.Info("Recieved work request")
				go func() {
					a.Logger.Info("Dispatching worker")
					worker := <-workerQueue
					worker <- work
				}()
			}
		}
	}()
}

func (a *App) Run() {
	a.Dispatcher()

	signal.Notify(a.Quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-a.Quit
		a.Logger.Info("Uptime bot is shutting down")
		time.Sleep(a.Wait)
		close(a.Done)
	}()

	go func() {
		for {
			a.Logger.Info("Scanning sites")
			a.PopulateSiteList()
			time.Sleep(a.Delay)
		}
	}()

	<-a.Done
	a.Logger.Info("Server stopped")
}
