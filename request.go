package main

import (
	"flag"
	"fmt"
	"net/http"
	urllib "net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	url          = "http://10.8.10.35"
	concurrency  = 100
	totalWorkers = 1000
	timeout      = 1
	start        = time.Now()
)

type Job struct {
	Id int
}

type Request struct {
	url          string
	concurrency  int
	totalWorkers int
	timeout      int

	failRequest   int64
	totalRequests int64
}

const concurrencySleep = 1000

func (r *Request) catchSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			fmt.Printf("\nTotal number of requests: %d\n", r.totalRequests)
			fmt.Printf("Number of failed requests: %d\n", r.failRequest)
			fmt.Printf("Interception success rate: %.2f\n", float64(r.failRequest)/float64(r.totalRequests))
			fmt.Println("HTTP requests take time：", time.Since(start))
			os.Exit(0)
		}
	}
}

func New(url string, concurrency, totalWorkers int) (*Request, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("%s", "The number of concurrences should be greater than 1.")
	}
	u, err := urllib.Parse(url)
	if err != nil || len(u.Host) == 0 {
		return nil, fmt.Errorf("please specify a correct url,%v", err)
	}
	return &Request{
		url:          url,
		concurrency:  concurrency,
		totalWorkers: totalWorkers,
	}, nil
}

func (r *Request) worker(id int, wg *sync.WaitGroup, jobChannel <-chan Job) {
	defer wg.Done()
	for job := range jobChannel {
		r.request(id, job)
	}
}

func (r *Request) request(workerId int, job Job) {
	var (
		err  error
		resp *http.Response
	)

	fmt.Printf("Worker #%d Running job #%d\n", workerId, job.Id)
	time.Sleep(concurrencySleep * time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(r.timeout) * time.Second,
	}

	if resp, err = client.Get(url); err != nil || resp.StatusCode != 200 {
		atomic.AddInt64(&r.failRequest, 1)
		//fmt.Printf("jobId #%d ----->  %s", job.Id, err)
	} else {
		//result, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(result))
	}
	atomic.AddInt64(&r.totalRequests, 1)
}

func (r *Request) Run() {
	var (
		wg   sync.WaitGroup
		jobs []Job
	)

	for i := 0; i < r.totalWorkers; i++ {
		jobs = append(jobs, Job{Id: i})
	}

	wg.Add(r.concurrency)

	jobChannel := make(chan Job)

	for i := 0; i < r.concurrency; i++ {
		go r.worker(i, &wg, jobChannel)
	}

	for _, job := range jobs {
		jobChannel <- job
	}
	close(jobChannel)
	wg.Wait()
}

func main() {
	flag.StringVar(&url, "u", url, "request url.")
	flag.IntVar(&concurrency, "c", concurrency, "number of concurrent requests (default 100, the program sleeps for 1 second after each concurrent request).")
	flag.IntVar(&totalWorkers, "w", totalWorkers, "maximum number of requests. (default 1000).")
	flag.IntVar(&timeout, "t", timeout, "request timeout setting. (default 1s, unit: seconds).")
	helpPtr := flag.Bool("h", false, "help.")
	flag.Usage = func() {
		_, _ = fmt.Fprint(os.Stderr,
			`Usage: ./request  OPTIONS [arg...]
       ./request [ -u http://10.8.10.35 -c 100 -w 10000 | -help ]
Options:
  -u          request url.
  -c          number of concurrent requests (default 100, the program sleeps for 1 second after each concurrent request).
  -w          maximum number of requests. (default 1000).
  -t          request timeout setting. (default 1s, unit: seconds).
  -h          help.
`)
	}
	flag.Parse()

	// show help
	if *helpPtr {
		flag.Usage()
		os.Exit(0)
	}

	r, err := New(url, concurrency, totalWorkers)
	if err != nil {
		panic(err)
	}

	go r.catchSignal()

	r.Run()
	fmt.Printf("\nTotal number of requests: %d\n", r.totalRequests)
	fmt.Printf("Number of failed requests: %d\n", r.failRequest)
	fmt.Printf("Interception success rate: %.2f\n", float64(r.failRequest)/float64(r.totalRequests))
	fmt.Println("HTTP requests take time：", time.Since(start))
}
