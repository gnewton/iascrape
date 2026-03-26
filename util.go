package iascrape

import (
	//"encoding/json"
	"errors"
	//"reflect"
	//"io"
	"log"
)

type StringFields struct {
	s    *[]string
	sRaw interface{}
}

func fixSearchItemStringFields(items []SearchItem) error {
	for i, _ := range items {
		item := &(items[i])

		sf := []StringFields{
			{&item.Title, item.Title_Raw},
			{&item.BackupLocation, item.BackupLocation_Raw},
			{&item.CurateNote, item.CurateNote_Raw},
			{&item.Curation, item.Curation_Raw},
			{&item.Format, item.Format_Raw},
			{&item.Date, item.Date_Raw},
		}

		err := fixStrings(sf)
		if err != nil {
			log.Printf("------ %#v\n", item)
			log.Println(err)
			return err
		}
	}
	return nil
}

func fixInts(ints [][]int, intsRaw []interface{}) error {
	for i := 0; i < len(ints); i++ {
		// ---------------
		intv, ok := intsRaw[i].(int)
		if ok {
			ints[i] = []int{intv}
		} else {
			inta, ok := intsRaw[i].([]int)
			if ok {
				ints[i] = inta
			} else {
				return errors.New("Non int or []int")
			}
		}
	}
	return nil
}

func fixStrings(sf []StringFields) error {
	for i := 0; i < len(sf); i++ {
		if sf[i].sRaw != nil {
			if v, ok := sf[i].sRaw.(string); ok {
				*sf[i].s = []string{v}
			} else {
				if inter, ok := sf[i].sRaw.([]interface{}); ok {
					*sf[i].s = make([]string, len(inter))

					for j := 0; j < len(inter); j++ {
						if v2, ok := inter[j].(string); ok {
							(*sf[i].s)[j] = v2
						}
					}
				}
			}
		}
	}
	return nil
}
