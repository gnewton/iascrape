package iascrape

import (
	"bytes"
	"compress/gzip"
	//"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"io"
	"io/ioutil"
	//"log"
	//"os"
	//"time"
)

var DBBucketName = "ia"

type Cache struct {
	db           *bolt.DB
	filename     string
	DBBucketName string
}

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

func (c *Cache) GetKey(url string) ([]byte, error) {
	var v []byte

	if err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DBBucketName))
		v = b.Get([]byte(url))
		return nil
	}); err != nil {
		return nil, err
	}

	if v != nil {
		var buf bytes.Buffer
		err := gunzipper2(&buf, v)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, nil
}

func (c *Cache) AddToCache(url string, body []byte) error {

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DBBucketName))

		var gzbuf bytes.Buffer
		gzipper(&gzbuf, []byte(body))

		return b.Put([]byte(url), gzbuf.Bytes())
	})
}

func gzipper(w io.Writer, data []byte) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	_, err := gw.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func gunzipper(data []byte) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	return err
}

func gunzipper2(w io.Writer, data []byte) error {
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
