package iascrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
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

func getUrlJSON(ctx context.Context, client *http.Client, url string, alternateKey string, results interface{}, cursor string, cache *Cache) error {
	if ctx == nil {
		return errors.New("Context is nil")
	}

	if url == "" {
		return errors.New("URL is empty string")
	}

	log.Println("Getting ", url)

	var key string
	var err error

	if alternateKey == "" {
		key = url
	} else {
		key = alternateKey
	}

	var body []byte
	if cache != nil {
		body, err = cache.GetKey(key)
		if err != nil {
			return err
		}
		log.Println("********* Cache hit")

	}

	if body == nil {
		body, err := getUrl(ctx, client, url)

		if err != nil {
			return err
		}
		if cache != nil {
			cache.AddToCache(key, body)
		}

		dec := json.NewDecoder(bytes.NewBuffer(body))

		return dec.Decode(results)
	}
	return nil
}

func getUrl(ctx context.Context, client *http.Client, urlString string) ([]byte, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	var res *http.Response

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
		return nil, errors.New("Request timedout")
	case <-ch:
		if err != nil {
			return nil, err
		}

		if res.StatusCode != 200 {
			body, err := io.ReadAll(res.Body)
			if err == nil {
				log.Println("Error. Response body:")
				log.Println("--------------------------------------------------------------")
				log.Println(string(body))
				log.Println("--------------------------------------------------------------")
			}
			return nil, fmt.Errorf("Failing http status code %d (!200)", res.StatusCode)
		}
		if err != nil {
			log.Println("Status code", res.StatusCode)
			log.Println(u)
			log.Println(err)
			return nil, err
		}
	}
	return io.ReadAll(res.Body)
}
