package iascrape

import (
	"bytes"
	"compress/gzip"
	//"errors"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"io"
	"io/ioutil"
	//"strings"
	//"log"
	//"os"
	//"time"
)

var DBBucketName = "ia"

// Cache stores the bbolt KV store information needed for the cache.
// It uses a single bucket to store all cache values

type Cache struct {
	db           *bolt.DB
	filename     string
	DBBucketName string
}

// NewCache opens an existing or creates a new boltdb at dbFileName.
// Returns a pointer to a Cache struct.
func NewCache(dbFileName string) (*Cache, error) {
	c := new(Cache)
	var err error

	c.db, err = bolt.Open(dbFileName, 0600, nil)
	if err != nil {
		return nil, err
	}

	return c, c.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DBBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

}

// Get uses the key argument to pull out an item from the cache (kv store)
// If successful, returns []byte, nil
// If unsuccessful, returns nil, nil
func (c *Cache) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("Empty URL string")
	}
	//

	var v []byte

	if err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DBBucketName))
		v = b.Get([]byte(key))
		return nil
	}); err != nil {
		return nil, err
	}

	if v != nil {
		var buf bytes.Buffer
		err := gunzipper(&buf, v)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, nil
}

func (c *Cache) Delete(key string) error {
	if key == "" {
		return errors.New("Empty URL string")
	}
	//

	var err error
	c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DBBucketName))
		err = b.Delete([]byte(key))
		return err
	})
	return err
}

func (c *Cache) Put(key string, body []byte) error {
	if key == "" {
		return errors.New("Empty URL string")
	}

	if len(body) == 0 {
		return errors.New("Empty value []byte")
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DBBucketName))

		var gzbuf bytes.Buffer
		gzipper(&gzbuf, []byte(body))

		return b.Put([]byte(key), gzbuf.Bytes())
	})
}

func gzipper(w io.Writer, data []byte) error {
	if w == nil {
		return errors.New("io.Writer is nil")
	}
	if data == nil {
		return errors.New("[]byte is nil")
	}
	if len(data) == 0 {
		return errors.New("[]bytes is zero length")
	}
	//

	gw := gzip.NewWriter(w)
	defer gw.Close()

	_, err := gw.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func gunzipper(w io.Writer, data []byte) error {
	if w == nil {
		return errors.New("io.Writer is nil")
	}
	if data == nil {
		return errors.New("[]byte is nil")
	}

	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
