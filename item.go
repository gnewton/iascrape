package iascrape

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	//"time"
)

// var ItemBaseUrl = "https://archive.org/metadata/"
var ItemBaseUrl = "http://archive.org/metadata/"

var IASCRAPE_DEBUG = false
var IASCRAPE_DEGUG_DEPTH = 0

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
	Segments         []string     `json:"-"`
	Segments_Raw     interface{}  `json:"segments"`
	Server           string       `json:"server"`
	Workable_Servers []string     `json:"workable_servers"`
	Uniq             int64        `json:"uniq"`
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
	Source_Raw                  interface{} `json:"source"`
	Subject_Raw                 interface{} `json:"subject"`
	Title_Raw                   interface{} `json:"title"`
	Uploader_Raw                interface{} `json:"uploader"`
	Year_Raw                    interface{} `json:"year"`
}

type ItemMetadata struct {
	ItemMetadata_Raw
	AddedDate               string   `json:"addeddate"`
	Collections             []string `json:"-"`
	collectionCatalogNumber []string `json:"-"`
	Condition               string   `json:"condition"`
	Contributor             string   `json:"contributor"`
	Creators                []string `json:"-"`
	Dates                   []string `json:"-"`
	Descriptions            []string `json:"-"`
	Genres                  []string `json:"-"`
	Identifier              string   `json:"identifier"`
	Keywords_CommaSeparated string   `json:"keywords"`
	Keywords                []string `json:"-"`
	Languages               []string `json:"-"`
	MediaType               string   `json:"media_type"`
	Notes                   []string `json:"-"`
	LicenseUrl              string   `json:"licenseurl"`
	PublicDate              string   `json:"publicdate"`
	Publishers              []string `json:"-"`
	PublisherCatalogNumbers []string `json:"-"`
	Scanners                []string `json:"-"`
	Source                  []string `json:"-"`
	Subjects                []string `json:"-"`
	Titles                  []string `json:"-"`
	Uploaders               []string `json:"-"`
	Years                   []string `json:"-"`
	CanonicalYear           int
}

type File struct {
	Format       string      `json:"format"`
	MD5          string      `json:"md5"`
	Name         string      `json:"name"`
	Size         string      `json:"size"`
	Title        string      `json:"title"`
	Original     []string    `json:"-"`
	Original_Raw interface{} `json:"original"`
	Length       string      `json:"length"`
	TrackOrder   int         `json:"-"` // This is not part of the JSON
}

type Role struct {
	Performer_Raw interface{} `json:"performer"`
	Performers    []string
}

func MakeMetadataItemFieldMap(md *ItemMetadata) map[string]*[]string {
	m := make(map[string]*[]string)

	m["creator"] = &md.Creators
	m["genre"] = &md.Genres
	m["keywords"] = &md.Keywords
	m["language"] = &md.Languages
	m["collection"] = &md.Collections
	m["subject"] = &md.Subjects
	m["title"] = &md.Titles
	return m
}

func GetItem(id string, client *http.Client, cache *Cache, verbose bool) (*ItemTopLevelMetadata, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty string")
	}

	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	url := ItemBaseUrl + id

	var item ItemTopLevelMetadata

	err, cacheHit := getUrlJSON(client, url, 6, id, &item, "", cache, verbose)
	if err != nil {
		return nil, err
	}

	// Delete, then re-get with cache turned off
	if len(item.Metadata.Identifier) == 0 {
		log.Println("#####$$$$$$$ URL:", url)
		log.Println("#####$$$$$$$ Cache:", cacheHit)
		log.Println("Missing identifier", url)

		err = cache.Delete(id)
		if err != nil {
			log.Println("Error deleting from cache id=", id)
			return nil, err
		}
		// Force to not use cache (nil cache)
		err, cacheHit := getUrlJSON(client, url, 6, id, &item, "", nil, verbose)
		if err != nil {
			return nil, err
		}
		if len(item.Metadata.Identifier) == 0 {
			log.Println("22 #####$$$$$$$ Cache:", cacheHit)
			log.Println("22 Missing identifier", url)
			return nil, errors.New("Identifier nil after second non-cache retrieval! ID=" + id + "      URL=" + url)
		} else {
			log.Println("OK", item.Metadata.Identifier)
		}

	}

	fixItemStringFields(&item)

	item.Metadata.CanonicalYear = makeYear(item.Metadata.Years, item.Metadata.Dates)

	return &item, nil
}

func fixItemStringFields(tm *ItemTopLevelMetadata) error {
	splitKeywordField(&tm.Metadata)

	sf := []StringFields{
		{&tm.Metadata.Source, tm.Metadata.Source_Raw},
		{&tm.Metadata.Collections, tm.Metadata.Collection_Raw},
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

	for i := 0; i < len(tm.Files); i++ {
		sff := StringFields{
			s:    &tm.Files[i].Original,
			sRaw: tm.Files[i].Original_Raw,
		}
		sf = append(sf, sff)
	}

	err := fixStrings(sf)
	if err != nil {
		log.Println(err)
		return err
	}
	splitSubjectField(&tm.Metadata)

	return nil
}

func isInt(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}

func makeYear(years []string, dates []string) int {
	year := ""
	date := ""
	if len(years) > 0 {
		year = years[0]
	}

	if len(dates) > 0 {
		date = dates[0]
	}

	if len(year) == 0 {
		if len(date) == 0 {
			return 0
		} else {
			return yearFromString(date)
		}
	}
	return yearFromString(year)

}

func yearFromString(s string) int {
	// 1984
	if len(s) == 4 && isInt(s) {
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0
		} else {
			return v
		}
	}

	// 1984-02
	// 1984-02-09
	if len(s) > 4 && isInt(s[0:4]) {
		v, err := strconv.Atoi(s[0:4])
		if err != nil {
			return 0
		} else {
			return v
		}
	}

	// 9/11/84
	parts := strings.Split(s, "/")
	if len(parts) == 3 {
		v, err := strconv.Atoi("19" + parts[2])
		if err != nil {
			return 0
		} else {
			return v
		}
	}

	if len(s) == 5 && s[0] == '~' {
		return yearFromString(s[1:])
	}

	return 0
}

// Some keyword fields are string (with single string), []string, and string with multiple values separated by a comma ","
func splitKeywordField(md *ItemMetadata) {
	md.Keywords = strings.Split(md.Keywords_CommaSeparated, ",")

	for i := 0; i < len(md.Keywords); i++ {
		md.Keywords[i] = strings.TrimSpace(md.Keywords[i])
	}
}

// Some subject fields are string (with single string), []string, and string with multiple values separated by a semicolon ";"
func splitSubjectField(md *ItemMetadata) {

	if len(md.Subjects) > 0 {
		md.Subjects = strings.Split(md.Subjects[0], ";")
	}

	for i := 0; i < len(md.Subjects); i++ {
		md.Subjects[i] = strings.TrimSpace(md.Subjects[i])
	}
}
