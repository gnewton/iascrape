package iascrape

import (
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	m.Run()
}

func Test_URL_getUrlJSON_NilClient(t *testing.T) {
	err, _ := getUrlJSON(nil, "http://example.com", 1, "", "", "", nil, false)

	if err == nil {
		t.Error("Should fail with nil client")
	}
}

func Test_URL_getUrlJSON_URLNoHttp(t *testing.T) {
	client := NewClient()

	err, _ := getUrlJSON(client, "example.com", 1, "", "", "", nil, false)

	if err == nil {
		t.Error("Should fail with nil client")
	}
}

func Test_URL_getUrlJSON_EmptyURL(t *testing.T) {
	client := NewClient()

	err, _ := getUrlJSON(client, "", 1, "", "", "", nil, false)

	if err == nil {
		t.Error("Should fail with nil client")
	}
}

func Test_URL_getUrlJSON_URLFailParse(t *testing.T) {
	client := NewClient()

	err, _ := getUrlJSON(client, "http://example.com hello doo", 1, "", "", "", nil, false)

	if err == nil {
		t.Error("Should fail unparseable URL")
	}
}

func Test_URL_getUrl_URLFailParse(t *testing.T) {
	client := NewClient()

	_, err := getUrl(client, "http://example.com hello doo", 1, 1)

	if err == nil {
		t.Error("Should fail unparseable URL")
	}
}

func Test_URL_getUrl_URLNoHttp(t *testing.T) {
	client := NewClient()

	_, err := getUrl(client, "example.com", 1, 1)

	if err == nil {
		t.Error("Should fail with nil client")
	}
}

func Test_URL_getUrl_NilClient(t *testing.T) {
	_, err := getUrl(nil, "http://example.com", 1, 1)

	if err == nil {
		t.Error("Should fail with nil client")
	}
}
