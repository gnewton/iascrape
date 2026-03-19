package iascrape

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Internet Archive Search api (scrape): https://archive.org/help/aboutsearch.htm
var IA_ScrapeBaseURL = "https://archive.org/services/search/v1/scrape?"

const MAX_RESULTS = 5000

type searchItems struct {
	Count          int    `json:"count"`
	Cursor         string `json:"cursor"`
	CursorPrevious string `json:"previous"`
	Items          []SearchItem
	Total          int64 `json:"total"`
}

type SearchItem struct {
	AddedDate              string      `json:"addeddate"`
	AvgRating_Raw          interface{} `json:"avg_rating"`
	AvgRating              []int
	BTIH                   string      `json:"btih"`
	BackupLocation_Raw     interface{} `json:"backup_location"`
	BackupLocation         []string
	Collection             []string    `json:"collection"`
	CollectionsOrdered     string      `json:"collections_ordered"`
	CurateDate             string      `json:"curatedate"`
	CurateNote_Raw         interface{} `json:"curatenote"`
	CurateNote             []string
	CurateState            string      `json:"curatestate"`
	Curation_Raw           interface{} `json:"curation"`
	Curation               []string
	Curator                string      `json:"curator"`
	Date_Raw               interface{} `json:"date"`
	Date                   []string
	Description            interface{} `json:"description"`
	Downloads              int         `json:"downloads"`
	ExternalMetadataUpdate string      `json:"external_metadata_update"`
	FilesCount             int         `json:"files_count"`
	Format_Raw             interface{} `json:"format"`
	Format                 []string
	Identifier             string      `json:"identifier"`
	IndexDate              string      `json:"indexdate"`
	ItemSize               int         `json:"item_size"`
	LicenseURL             string      `json:"licenseurl"`
	ListMemberships_Raw    interface{} `json:"list_memberships"`
	ListMemberships        []string
	MatchDateAoustid       string      `json:"match_date_acoustid"`
	MediaType              string      `json:"mediatype"`
	Month                  int         `json:"month"`
	NoArchiveTorrent       string      `json:"noarchivetorrent"`
	NumFavorites           int         `json:"num_favorites"`
	OaiUpdateDate_Raw      interface{} `json:"oai_updatedate"`
	OaiUpdateDate          []string
	PrimaryCollection      string      `json:"primary_collection"`
	PublicDate             string      `json:"publicdate"`
	ReportedServer         string      `json:"reported_server"`
	ReviewBody_Raw         interface{} `json:"reviewbody"`
	ReviewBody             []string
	ReviewData             []string    `json:"review_data"`
	Reviewer_Raw           interface{} `json:"reviewer"`
	Reviewer               []string
	ReviewerItemName_Raw   interface{} `json:"reviewer_itemname"`
	ReviewerItemname       []string
	Scanner_Raw            interface{} `json:"scanner"`
	Scanner                []string
	Subject_Raw            interface{} `json:"subject"`
	Subject                []string
	SubjectCount           int         `json:"subject_count"`
	Stars_Raw              interface{} `json:"stars"`
	Stars                  []int
	Title_Raw              interface{} `json:"title"`
	Title                  []string
	Week                   int         `json:"week"`
	Year_Raw               interface{} `json:"year"`
	Year                   []int
}

type Search struct {
	Query        string
	cursor       string
	resultsCount int64
	MaxResults   int64
	ChunkSize    int
	Client       *http.Client
	done         bool
	pause        time.Duration
	Retries      int
}

func (s *Search) Total() (int64, error) {
	if s.Query == "" {
		return 0, errors.New("Query cannot be empty string")
	}

	url := IA_ScrapeBaseURL + s.Query + "&total_only=true"

	var results searchItems
	var err error

	err = getUrlJSON(s.Client, url, 5, "", &results, s.cursor, nil)
	if err != nil {
		return 0, err
	}
	return results.Total, nil
}

func (s *Search) Execute() ([]SearchItem, error) {

	if s.MaxResults < 100 {
		return nil, fmt.Errorf("Requested num results must be > 100")
	}

	if s.ChunkSize > 5000 {
		return nil, fmt.Errorf("ChunkSize number of results requested exceeded")
	}

	if s.done {
		return nil, nil
	}

	thisQuery := s.Query

	if s.ChunkSize != 0 {
		thisQuery = thisQuery + "&count=" + strconv.Itoa(s.ChunkSize)
	}

	if s.cursor != "" {
		thisQuery = thisQuery + "&cursor=" + s.cursor
	}

	var tmpItems searchItems
	url := IA_ScrapeBaseURL + thisQuery

	log.Println("search", url)

	err := getUrlJSON(s.Client, url, 6, "", &tmpItems, s.cursor, nil)
	if err != nil {
		return nil, err
	}

	if len(tmpItems.Items) == 0 {
		return nil, nil
	}

	s.resultsCount = s.resultsCount + int64(len(tmpItems.Items))

	if s.resultsCount >= s.MaxResults {
		s.done = true
	}

	err = fixSearchItemStringFields(tmpItems.Items)
	if err != nil {
		return nil, err
	}
	s.cursor = tmpItems.Cursor

	if s.cursor == "" {
		s.done = true
	}

	return tmpItems.Items, nil
}
