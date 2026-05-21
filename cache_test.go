package iascrape

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func Test_NewCache_EmptyStringFileArg(t *testing.T) {
	_, err := NewCache("")

	if err == nil {
		t.Error(err)
	} else {
		log.Println(err)
	}
}

func Test_NewCache_NoPermissionsFileLocation(t *testing.T) {
	_, err := NewCache("/foo")

	if err == nil {
		t.Error(err)
	} else {
		log.Println(err)
	}
}

func Test_NewCache_FileLocationIsDirectory(t *testing.T) {
	_, err := NewCache("/tmp")

	if err == nil {
		t.Error(err)
	} else {
		log.Println(err)
	}
}

func Test_CacheGet_KeyEmptyString(t *testing.T) {
	c, err := makeTmpCache()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(c.filename)
	//

	_, err = c.Get("")

	if err == nil {
		t.Error("Should fail on empty string key")
	}

}

func Test_CacheDeleteKey_KeyEmptyString(t *testing.T) {
	c, err := makeTmpCache()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(c.filename)

	err = c.Delete("")
	if err == nil {
		t.Error("Emptry string should be error")
	}
}

func Test_CachePutKey_KeyEmptyString(t *testing.T) {
	c, err := makeTmpCache()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(c.filename)

	err = c.Put("", []byte("valueBar"))
	if err == nil {
		t.Error("Emptry string should be error")
	}
}

func Test_CachePutKey_ValueEmptyString(t *testing.T) {
	c, err := makeTmpCache()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(c.filename)

	err = c.Put("keyFoo", []byte(""))
	if err == nil {
		t.Error("Emptry string should be error")
	}
}

func Test_Cachegzipper_NilWriter(t *testing.T) {
	err := gzipper(nil, []byte("bar"))

	if err == nil {
		t.Error("Should fail with nil io.Writer")
	}
}

func Test_Cachegzipper_EmptyData(t *testing.T) {
	var w bytes.Buffer
	err := gzipper(&w, []byte(""))

	if err == nil {
		t.Error("Should fail with nil io.Writer")
	}
}

func Test_Cacheungzipper_DataNil(t *testing.T) {
	var w bytes.Buffer
	err := gunzipper(&w, nil)

	if err == nil {
		t.Error("Should fail with nil io.Writer")
	}
}

func Test_Cacheungzipper_WriterNil(t *testing.T) {

	err := gunzipper(nil, []byte(""))

	if err == nil {
		t.Error("Should fail with nil io.Writer")
	}
}

// /////////////////////////////////////////////////
// Util
func makeTmpFile() (*os.File, error) {
	return os.CreateTemp("", "iascrape_cache_")
}

func makeTmpCache() (*Cache, error) {
	f, err := makeTmpFile()
	if err != nil {
		return nil, err
	}
	c, err := NewCache(f.Name())
	if err != nil {
		return nil, err
	}

	return c, nil

}
