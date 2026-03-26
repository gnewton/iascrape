package iascrape

import (
	"errors"
	"log"
	"net/http"
	"strings"
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
	CollectionCatalogNumber_Raw interface{} `json:"collection-catalog-number"`
	Collection_Raw              interface{} `json:"collection"`
	Creator_Raw                 interface{} `json:"creator"`
	Date_Raw                    interface{} `json:"date"`
	Description_Raw             interface{} `json:"description"`
	Genre_Raw                   interface{} `json:"genre"`
	Language_Raw                interface{} `json:"language"`
	Notes_Raw                   interface{} `json:"notes"`
	PublisherCatalogNumber_Raw  interface{} `json:"publisher-catalog-number"`
	Publisher_Raw               interface{} `json:"publisher"`
	Scanner_Raw                 interface{} `json:"scanner"`
	Subject_Raw                 interface{} `json:"subject"`
	Title_Raw                   interface{} `json:"title"`
	Uploader_Raw                interface{} `json:"uploader"`
	Year_Raw                    interface{} `json:"year"`
}

type ItemMetadata struct {
	ItemMetadata_Raw
	AddedDate               string `json:"addeddate"`
	Collections             []string
	collectionCatalogNumber []string
	Condition               string `json:"condition"`
	Contributor             string `json:"contributor"`
	Creators                []string
	Dates                   []string
	Descriptions            []string
	Genres                  []string
	Identifier              string `json:"identifier"`
	Keywords_CommaSeparated string `json:"keywords"`
	Keywords                []string
	Languages               []string
	MediaType               string `json:"media_type"`
	Notes                   []string
	LicenseUrl              string `json:"licenseurl"`
	PublicDate              string `json:"publicdate"`
	Publishers              []string
	PublisherCatalogNumbers []string
	Scanners                []string
	Subjects                []string
	Titles                  []string
	Uploaders               []string
	Years                   []string
}

type File struct {
	Format string `json:"format"`
	MD5    string `json:"md5"`
	Name   string `json:"name"`
	Size   string `json:"size"`
	Title  string `json:"title"`
}

type Role struct {
	Performer_Raw interface{} `json:"performer"`
	Performers    []string
}

func MakeMetadataItemFieldMap(md *ItemMetadata) map[string]*[]string {
	m := make(map[string]*[]string)

	//m["collection"] = &md.Collection
	m["creator"] = &md.Creators
	m["genre"] = &md.Genres
	m["keywords"] = &md.Keywords
	m["language"] = &md.Languages
	m["subject"] = &md.Subjects
	return m
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
		log.Println("ID=", id)
		return nil, err
	}

	fixItemStringFields(&item)

	return &item, nil
}

func fixItemStringFields(tm *ItemTopLevelMetadata) error {
	fixKeywordField(&tm.Metadata)

	sf := []StringFields{
		//{&tm.Metadata.Collection, tm.Metadata.Collection_Raw},
		{&tm.Metadata.collectionCatalogNumber, tm.Metadata.CollectionCatalogNumber_Raw},
		{&tm.Metadata.Creators, tm.Metadata.Creator_Raw},
		{&tm.Metadata.Dates, tm.Metadata.Date_Raw},
		{&tm.Metadata.Descriptions, tm.Metadata.Description_Raw},
		{&tm.Metadata.Languages, tm.Metadata.Language_Raw},
		{&tm.Metadata.Scanners, tm.Metadata.Scanner_Raw},
		{&tm.Metadata.Subjects, tm.Metadata.Subject_Raw},
		{&tm.Metadata.Titles, tm.Metadata.Title_Raw},
		{&tm.Metadata.Uploaders, tm.Metadata.Uploader_Raw},
		{&tm.Metadata.Years, tm.Metadata.Year_Raw},
		{&tm.Roles.Performers, tm.Roles.Performer_Raw},
		{&tm.Segments, tm.Segments_Raw},
	}

	err := fixStrings(sf)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func fixKeywordField(md *ItemMetadata) {
	md.Keywords = strings.Split(md.Keywords_CommaSeparated, ",")

	for i := 0; i < len(md.Keywords); i++ {
		md.Keywords[i] = strings.TrimSpace(md.Keywords[i])
	}

	// var tmp []string
	// for i := 0; i < len(md.Subjects); i++ {
	// 	z := md.Subjects[i]
	// 	md.Subjects[i] = strings.TrimSpace(md.Subjects[i])
	// }
	// md.Subjects = strings.Split(md.Subjects_CommaSeparated, ",")

}
