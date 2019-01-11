package worker

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

var (
	WorkQueue = make(chan WorkRequest, QueueSize)
	QueueSize = 100
)

type WorkRequest struct {
	Email  string
	Name   string
	Status int
	Type   string
	URL    string
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
}

func NewWorker(id int, workerQueue chan chan WorkRequest) Worker {
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
	}
	return worker
}

func NewWorkRequest(status int, email, name, wtype, url string) WorkRequest {
	work := WorkRequest{
		Email:  email,
		Name:   name,
		Status: status,
		Type:   wtype,
		URL:    url,
	}
	return work
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				switch work.Type {
				case "request":
					log.WithFields(logrus.Fields{
						"WorkerId": w.ID,
						"Name":     work.Name,
						"URL":      work.URL,
					}).Info("request made")
					work.MakeRequest()
				case "response":
					log.Info("Checking response")
					work.ProcessResponse()
				default:
					log.Warn("Invalid Work Request Type")
				}
			}
		}
	}()
}

var clientPool = sync.Pool{
	New: func() interface{} {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				DualStack: true,
				KeepAlive: time.Second * 10,
				Timeout:   time.Second * 10,
			}).DialContext,
		}
		client := &http.Client{
			Timeout:   time.Second * 10,
			Transport: tr,
		}
		return client
	},
}

func (wr WorkRequest) MakeRequest() {
	client := clientPool.Get().(*http.Client)
	defer clientPool.Put(client)

	log.Info("Making request to: ", wr.URL)
	req, err := http.NewRequest(http.MethodHead, wr.URL, nil)
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

	wr.Status = resp.StatusCode
	wr.Type = "response"
	log.Info("struct data:     ", wr)
	WorkQueue <- wr
}

func (wr WorkRequest) ProcessResponse() {
	if wr.Status > 399 {
		log.WithFields(logrus.Fields{
			"URL":    wr.URL,
			"Status": wr.Status,
			"Email":  wr.Email,
		}).Warn("BAD REQUEST SEND EMAIL")
		wr.SendEmail()
	} else {
		log.WithFields(logrus.Fields{
			"URL":    wr.URL,
			"Status": wr.Status,
			"Email":  wr.Email,
		}).Info("URL OK")
	}
}

func (wr WorkRequest) SendEmail() {
	from := mail.NewEmail("Uptime status", os.Getenv("UPTIME_EMAIL"))
	subject := "Website Uptime Warning"
	to := mail.NewEmail("Uptime error email", wr.Email)
	plainTextContent := "Uptime error for website"
	htmlContent := "The website <a href=\"" + wr.URL + "\">" + wr.URL + "</a>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	resp, err := client.Send(message)
	if err != nil {
		log.Warn("sendgrid error: ", err)
		return
	}

	log.Info(resp.StatusCode)
	log.Info(resp.Body)
	log.Info(resp.Headers)
}

func ValidateURL(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		log.Fatal("cannot parse URL")
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	return u.String()
}
