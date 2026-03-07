package iascrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	//"net"
	"context"
	"net/http"
	"time"
)

var min time.Duration = 99999999999999
var max time.Duration = 0
var n int64 = 0
var total time.Duration = 0

var backOff time.Duration = 3 * time.Second

type RequestStats struct {
	lastTime time.Duration
	min      time.Duration
	max      time.Duration
	n        int64
}

type Backoff func(*RequestStats) time.Duration

func NewClient() *http.Client {

	transport := &http.Transport{
		DisableCompression:  false,
		DisableKeepAlives:   false,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		// Dial: (&net.Dialer{
		// 	Timeout:   30 * time.Second,
		// 	KeepAlive: 30 * time.Second,
		// }).Dial,
		// TLSHandshakeTimeout:   15 * time.Second,
		// ResponseHeaderTimeout: 15 * time.Second,
		// ExpectContinueTimeout: 5 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   120 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Println("VIA", len(via), via)
			// Custom redirect handling
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
}

func getUrlJSON(ctx context.Context, client *http.Client, url string, alternateKey string, items interface{}, cursor string, cache *Cache) (time.Duration, error) {

	log.Println("Getting ", url)
	var body []byte
	var err error

	var key string
	if alternateKey == "" {
		key = url
	} else {
		key = alternateKey
	}

	if cache != nil {
		body, err = cache.GetKey(key)
		if err != nil {
			return 0, err
		}
		log.Println("********* Cache hit")

	}

	var since time.Duration
	if body == nil {
		startTime := time.Now()
		body, err := getUrl(ctx, client, url)
		since = time.Since(startTime)

		if err != nil {
			return 0, err
		}
		if cache != nil {
			cache.AddToCache(key, body)
		}

		dec := json.NewDecoder(bytes.NewBuffer(body))

		return since, dec.Decode(items)
	}
	log.Println("DDDDDDDDDDDDDDDDDDDDDDDDD")
	return 0, nil
}

func getUrl(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	var res *http.Response

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Error: Client fail: %s\n", err)
		return nil, err
	}

	ch := make(chan struct{})
	go func() {
		log.Println("URL start")
		res, err = client.Do(req)
		log.Println("URL end")
		ch <- struct{}{}
	}()

	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return nil, err
	}

	select {
	case <-ctx.Done():
		log.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
		return nil, errors.New("Request timedout")
	case <-ch:
		log.Println("BBBBBBBBBBBBBBBBBBBBBBBBb")
		if err != nil {
			return nil, err
		}
		log.Println("CCCCCCCCCCCCCCCCCCCCCCCCCCc")
		// n++
		// total += since
		// if since < min {
		// 	min = since
		// }

		// if since > max {
		// 	max = since
		// }

		// log.Println(since, min, max, time.Duration(int64(total)/n))

		// if since > time.Duration(int64(float64(int64(total))/float64(n)*3.0)) || since > 5*time.Second {
		// 	// Backoff
		// 	if backOff < 30*time.Second {
		// 		backOff = backOff + time.Second + time.Second + time.Second + time.Second
		// 	}
		// 	log.Println(backOff, "getUrlJSON - BACKOFF $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
		// 	log.Println(backOff + time.Second + time.Second + time.Second)
		// 	time.Sleep(backOff + time.Second + time.Second + time.Second)
		// } else {
		// 	if backOff > 0 {
		// 		backOff = backOff - time.Second - time.Second
		// 	}
		// }

		if res.StatusCode != 200 {
			body, err := io.ReadAll(res.Body)
			if err == nil {
				log.Println("Error. Response body:")
				log.Println("--------------------------------------------------------------")
				log.Println(string(body))
				log.Println("--------------------------------------------------------------")
			}
			return nil, fmt.Errorf("Failing http code %d (!200)", res.StatusCode)
		}
		if err != nil {
			log.Println("Status code", res.StatusCode)
			log.Println(url)
			log.Println(err)
			return nil, err
		}
	}

	return io.ReadAll(res.Body)

}
