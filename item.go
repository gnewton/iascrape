package iascrape

import (
	"errors"
	"log"
	"net/http"

	//"time"

	"context"
)

// var ItemBaseUrl = "https://archive.org/metadata/"
var ItemBaseUrl = "http://archive.org/metadata/"

type ItemTopLevelMetadata struct {
	Created         int64        `json:"created"`
	D1              string       `json:"d1"`
	Date            string       `json:"date"`
	Dir             string       `json:"dir"`
	Files           []File       `json:"files"`
	ItemLastUpdated int          `json:"item_last_updated"`
	ItemSize        int64        `json:"item_size"`
	Metadata        ItemMetadata `json:"metadata"`
	Roles           Role         `json:"roles"`
	Segments        []string
	Segments_Raw    interface{} `json:"segments"`
}

type ItemMetadata struct {
	AddedDate                   string      `json:"addeddate"`
	CollectionCatalogNumber_Raw interface{} `json:"collection-catalog-number"`
	CollectionCatalogNumber     []string
	Creator                     []string
	Creator_Raw                 interface{} `json:"creator"`
	Date                        []string
	Date_Raw                    interface{} `json:"date"`
	Identifier                  string      `json:"identifier"`
	Keywords                    string      `json:"keywords"`
	Language                    []string
	Language_Raw                interface{} `json:"language"`
	MediaType                   string      `json:"media_type"`
	PublicDate                  string      `json:"publicdate"`
	Scanner                     []string
	Scanner_Raw                 interface{} `json:"scanner"`
	Subject                     []string
	Subject_Raw                 interface{} `json:"subject"`
	Title                       []string
	Title_Raw                   interface{} `json:"title"`
	Uploader_Raw                interface{}      `json:"uploader"`
	Uploader                    []string     
	Year                        []string
	Year_Raw                    interface{} `json:"year"`
}

type File struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Title  string `json:"title"`
	Size   string `json:"size"`
}

type Role struct {
	Performer_Raw interface{} `json:"performer"`
	Performer     []string
}

func GetItem(ctx context.Context, id string, client *http.Client, cache *Cache) (*ItemTopLevelMetadata, error) {
	if ctx == nil {
		return nil, errors.New("Context is nil")
	}
	if id == "" {
		return nil, errors.New("id cannot be empty string")
	}

	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	url := ItemBaseUrl + id

	var item ItemTopLevelMetadata
	//err := getUrlJSON(ctx, client, url, id, &item, "", cache)
	err := getUrlJSON2(client, url, 6, id, &item, "", cache)
	if err != nil {
		return nil, err
	}

	fixItemStringFields(&item)

	return &item, nil
}

func fixItemStringFields(tm *ItemTopLevelMetadata) error {
	sf := []StringFields{
		{&tm.Metadata.Uploader, tm.Metadata.Uploader_Raw},
		{&tm.Segments, tm.Segments_Raw},
		{&tm.Metadata.Subject, tm.Metadata.Subject_Raw},
		{&tm.Metadata.Creator, tm.Metadata.Creator_Raw},
		{&tm.Metadata.Title, tm.Metadata.Title_Raw},
		{&tm.Metadata.Year, tm.Metadata.Year_Raw},
		{&tm.Metadata.Language, tm.Metadata.Language_Raw},
		{&tm.Metadata.Scanner, tm.Metadata.Scanner_Raw},
		{&tm.Metadata.Date, tm.Metadata.Date_Raw},
		{&tm.Metadata.CollectionCatalogNumber, tm.Metadata.CollectionCatalogNumber_Raw},
		{&tm.Roles.Performer, tm.Roles.Performer_Raw},
	}

	err := fixStrings(sf)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
