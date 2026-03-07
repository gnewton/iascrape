package iascrape

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var DBBucketName = "ia"

type Cache struct {
	db           *bolt.DB
	filename     string
	DBBucketName string
	KeepForever  bool
}

func (c *Cache) InitializeCache(dbFileName string) error {
	log.Println("InitializeCache", dbFileName)
	if c.db != nil {
		return errors.New("DB is not nil; only run initializeCache() once!")
	}

	var err error
	log.Println(c.KeepForever)

	if !c.KeepForever {
		log.Println("Delewting cache db", dbFileName)
		fileInfo, err := os.Stat(dbFileName)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		// Gives the modification time
		if fileInfo != nil {
			modificationTime := fileInfo.ModTime()
			if time.Since(modificationTime) > time.Hour {
				err = os.Remove(dbFileName)
				if err != nil {
					log.Println("Unable to delete file:", dbFileName)
					return err
				}
			}

		}
	}

	c.db, err = bolt.Open(dbFileName, 0600, nil)
	if err != nil {
		return err
	}

	return c.db.Update(func(tx *bolt.Tx) error {
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
	} else {
		log.Println("Cache miss")
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
