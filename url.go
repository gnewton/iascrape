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
	//"context"
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

func getUrlJSON2(client *http.Client, urlString string, retry int, alternateKey string, results interface{}, cursor string, cache *Cache) error {
	if urlString == "" {
		return errors.New("URL is empty string")
	}

	_, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	var key string

	if alternateKey == "" {
		key = urlString
	} else {
		key = alternateKey
	}

	var body []byte
	if cache != nil {
		body, err = cache.GetKey(key)
		if err != nil {
			return err
		}
	}

	if body == nil {
		body, err := getUrl2(client, urlString, retry, time.Second*5)

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

func getUrl2(client *http.Client, u string, retry int, delay time.Duration) ([]byte, error) {
	log.Println("Getting ", u)
	var err error

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		fmt.Printf("Error: Client fail: %s\n", err)
		return nil, err
	}

	res, err := client.Do(req)

	if err != nil {
		if retry == 0 {
			fmt.Printf("client: error making http request: %s\n", err)
			return nil, err
		} else {
			log.Println("getUrl2: recurse", retry-1, delay*2, "   ==================================")
			return getUrl2(client, u, retry-1, delay*2)
		}
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
	return io.ReadAll(res.Body)
}
