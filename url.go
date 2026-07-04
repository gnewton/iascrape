package iascrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	//"net"
	//"context"
	"net/http"
	"runtime"
	"time"
)

const UserAgent = runtime.GOOS + "; iascrape; github.com/gnewton/iascrape"

var min time.Duration = 99999999999999
var max time.Duration = 0
var n int64 = 0
var total time.Duration = 0

var cacheHits int64 = 0
var cacheMisses int64 = 0

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
			for i := 0; i < len(via); i++ {
				log.Println(i, ".......", via[i].URL)
			}
			// Custom redirect handling
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
}

func getUrlJSON(client *http.Client, urlString string, retry int, alternateKey string, results interface{}, cursor string, cache *Cache, verbose bool) (error, bool) {
	if client == nil {
		return errors.New("client *http.Client is nil"), false
	}

	if hasHttpPrefix(urlString) {
		return errors.New("URL invalid: does not start with http:// or https://"), false
	}

	_, err := url.Parse(urlString)
	if err != nil {
		return err, false
	}

	var key string

	if alternateKey == "" {
		key = urlString
	} else {
		key = alternateKey
	}

	cacheHit := false
	var body []byte
	if cache != nil {
		body, err = cache.Get(key)
		if err != nil {
			return err, false
		}
		if verbose && body != nil {
			//log.Println("Cache hit")
		}
		if body != nil {
			cacheHit = true
			cacheHits += 1
		} else {
			cacheMisses += 1
		}
	}

	if body == nil {
		body, err = getUrl(client, urlString, retry, time.Second*5)
		//log.Println("Cache miss", urlString)
		if err != nil {
			return err, false
		}
		if cache != nil {
			cache.Put(key, body)
		}
	}

	if len(body) <= 2 {
		log.Println(string(body))
		return errors.New("Error: Returns empty JSON: " + urlString), false
	}

	// This is a hack. IA does not set the 	Header "content-type"
	if body[0] != '{' {
		log.Println("Content is not JSON", urlString)
		log.Fatal(string(body))
	}

	dec := json.NewDecoder(bytes.NewBuffer(body))

	err = dec.Decode(results)
	if err != nil {
		log.Println("Error decoding JSON for URL:", urlString)
		log.Println(string(body))
	}

	return err, cacheHit
}

func CacheStats() (int64, int64) {
	return cacheHits, cacheMisses
}

func hasHttpPrefix(u string) bool {
	return !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://")
}

func getUrl(client *http.Client, u string, retry int, delay time.Duration) ([]byte, error) {
	if client == nil {
		return nil, errors.New("client *http.Client is nil")
	}
	if hasHttpPrefix(u) {
		return nil, errors.New("URL invalid: does not start with http:// or https://")
	}

	var err error

	_, err = url.Parse(u)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		log.Printf("Error: Client fail: %s\n", err)
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := client.Do(req)

	if err != nil {
		if retry == 0 {
			log.Printf("client: error making http request: %s\n", err)
			return nil, err
		} else {
			log.Println("getUrl2: recurse", retry-1, delay*2, "   ==================================")
			return getUrl(client, u, retry-1, delay*2)
		}
	}

	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err == nil {
			log.Println("Error for url:", u)
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

	 if !headerContains(res.Header["Content-Type"], "application/json") {
	 	return nil, errors.New("Error: Content is not json")
	 }
	// for k, v := range res.Header {
	// 	log.Println(k, ":", v)
	// }
	return io.ReadAll(res.Body)
}

// Need to support both:
// Content-Type:[application/json; charset=UTF-8]
// Content-Type:[application/json]
func headerContains(header []string, s string) bool {
	if header == nil {
		return true
	}
	for i := 0; i < len(header); i++ {
		//if header[i] ==s {
		if strings.HasPrefix(header[i],s) {
			return true
		}
	}
	return false
}

func HeadUrl(client *http.Client, u string, retry int, delay time.Duration) error {

	var err error

	req, err := http.NewRequest(http.MethodHead, u, nil)
	if err != nil {
		fmt.Printf("Error: Client fail: %s\n", err)
		return err
	}

	res, err := client.Do(req)

	if err != nil {
		if retry == 0 {
			log.Printf("client: error making http %s request: %s\n", u, err)
			return err
		} else {
			log.Println("getUrl2: recurse", retry-1, delay*2, "   ==================================")
			return HeadUrl(client, u, retry-1, delay*2)
		}
	}

	log.Println("HeadURL", res.StatusCode, u)
	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err == nil {
			log.Println("Error for url:", u)
			log.Println("Error. Response body:")
			log.Println("--------------------------------------------------------------")
			log.Println(string(body))
			log.Println("--------------------------------------------------------------")
		}
		return fmt.Errorf("Failing http status code %d (!200)", res.StatusCode)
	}
	if err != nil {
		log.Println("Status code", res.StatusCode)
		log.Println(u)
		log.Println(err)
		return nil
	}
	return nil
}
