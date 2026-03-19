package iascrape

import (
	"errors"
	"log"
	"net/http"
	//"time"
)

// var ItemBaseUrl = "https://archive.org/metadata/"
var ItemBaseUrl = "http://archive.org/metadata/"

type ItemTopLevelMetadata struct {
	Created          int64        `json:"created"`
	D1               string       `json:"d1"`
	Date             string       `json:"date"`
	Dir              string       `json:"dir"`
	Files            []File       `json:"files"`
	Files_Count      int32        `json:"files_count"`
	ItemLastUpdated  int64        `json:"item_last_updated"`
	ItemSize         int64        `json:"item_size"`
	Metadata         ItemMetadata `json:"metadata"`
	Roles            Role         `json:"roles"`
	Segments         []string
	Segments_Raw     interface{} `json:"segments"`
	Server           string      `json:"server"`
	Workable_Servers []string    `json:"workable_servers"`
	Uniq             int64       `json:"uniq"`
}

type ItemMetadata_Raw struct {
	Collection_Raw              interface{} `json:"collection"`
	Date_Raw                    interface{} `json:"date"`
	Creator_Raw                 interface{} `json:"creator"`
	Description_Raw             interface{} `json:"description"`
	Genre_Raw                   interface{} `json:"genre"`
	CollectionCatalogNumber_Raw interface{} `json:"collection-catalog-number"`
	Language_Raw                interface{} `json:"language"`
	Notes_Raw                   interface{} `json:"notes"`
	Publisher_Raw               interface{} `json:"publisher"`
	Scanner_Raw                 interface{} `json:"scanner"`
	PublisherCatalogNumber_Raw  interface{} `json:"publisher-catalog-number"`
	Subject_Raw                 interface{} `json:"subject"`
	Title_Raw                   interface{} `json:"title"`
	Uploader_Raw                interface{} `json:"uploader"`
	Year_Raw                    interface{} `json:"year"`
}

type ItemMetadata struct {
	ItemMetadata_Raw
	AddedDate               string `json:"addeddate"`
	Collection              []string
	CollectionCatalogNumber []string
	Condition               string `json:"condition"`
	Contributor             string `json:"contributor"`
	Creator                 []string
	Date                    []string
	Description             []string
	Genre                   []string
	Identifier              string `json:"identifier"`
	Keywords                string `json:"keywords"`
	Language                []string
	MediaType               string `json:"media_type"`
	Notes                   []string
	LicenseUrl              string `json:"licenseurl"`
	PublicDate              string `json:"publicdate"`
	Publisher               []string
	PublisherCatalogNumber  []string
	Scanner                 []string
	Subject                 []string
	Title                   []string
	Uploader                []string
	Year                    []string
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

func GetItem(id string, client *http.Client, cache *Cache) (*ItemTopLevelMetadata, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty string")
	}

	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	url := ItemBaseUrl + id

	var item ItemTopLevelMetadata

	err := getUrlJSON(client, url, 6, id, &item, "", cache)
	if err != nil {
		return nil, err
	}

	fixItemStringFields(&item)

	return &item, nil
}

func fixItemStringFields(tm *ItemTopLevelMetadata) error {
	sf := []StringFields{
		{&tm.Metadata.Description, tm.Metadata.Description_Raw},
		{&tm.Metadata.Uploader, tm.Metadata.Uploader_Raw},
		{&tm.Segments, tm.Segments_Raw},
		{&tm.Metadata.Subject, tm.Metadata.Subject_Raw},
		{&tm.Metadata.Collection, tm.Metadata.Collection_Raw},
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
