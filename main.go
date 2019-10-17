package main

import (
	"bytes"
	"context"
	"encoding/json"
	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

/**
see http://www.golangprograms.com/remove-duplicate-values-from-slice.html
*/
func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

/**
prepare a string->string map that will be serialized to the json search document
and serialize it
*/
func make_query(environment string) bytes.Buffer {
	var buf bytes.Buffer

	log.Printf("Querying for NotificationException in fields.type==vidispine and fields.environment==%s", environment)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"match": map[string]interface{}{
						"class": "com.vidispine.notifications.NotificationException",
					},
				},
				"filter": []map[string]interface{}{
					//{"term": map[string]interface{}{
					//  "fields.type":"vidispine",
					//}},
					{"match": map[string]interface{}{
						"fields.environment": environment,
					}},
				},
			},
		},
	}

	encodeErr := json.NewEncoder(&buf).Encode(query)
	if encodeErr != nil {
		log.Fatalf("Could not encode query: %s", encodeErr)
	}
	return buf
}

/**
deserialize a returned buffer to a string->any map
*/
func generic_decode(reader io.ReadCloser) (*map[string]interface{}, error) {
	var e map[string]interface{}
	err := json.NewDecoder(reader).Decode(&e)

	if err == nil {
		return &e, nil
	} else {
		return nil, err
	}
}

func cleanup(esclient *elasticsearch6.Client, scrollId *string) error {
	_, err := esclient.ClearScroll(
		esclient.ClearScroll.WithContext(context.Background()),
		esclient.ClearScroll.WithScrollID(*scrollId),
	)
	return err
}

func find_records(esclient *elasticsearch6.Client, indexName string, queryBuffer bytes.Buffer, scrollId *string, offset int, limit int) (*[]Record, *string, error) {
	var result *esapi.Response
	var err error
	if scrollId == nil {
		result, err = esclient.Search(
			esclient.Search.WithContext(context.Background()),
			esclient.Search.WithIndex(indexName),
			esclient.Search.WithBody(&queryBuffer),
			esclient.Search.WithScroll(time.Duration(15)*time.Minute),
			esclient.Search.WithTrackTotalHits(true),
			esclient.Search.WithFrom(offset),
			esclient.Search.WithSize(limit),
		)
	} else {
		result, err = esclient.Scroll(
			esclient.Scroll.WithContext(context.Background()),
			esclient.Scroll.WithScrollID(*scrollId),
			esclient.Scroll.WithScroll(time.Duration(15)*time.Minute),
		)
	}

	if err != nil {
		return nil, nil, err
	}
	defer result.Body.Close()

	if result.IsError() {
		e, err := generic_decode(result.Body)
		if err != nil {
			log.Fatalf("Error parsing response body: %s", err)
		} else {
			log.Fatalf("ES reported error: %#v", e)
		}
	}

	var resp Response
	decodeErr := json.NewDecoder(result.Body).Decode(&resp)
	if decodeErr != nil {
		log.Fatalf("Error parsing response body: %s", decodeErr)
	}

	log.Printf("Got %d results: ", resp.Hits.Total)
	var rtn []Record
	for _, h := range resp.Hits.Hits {
		rtn = append(rtn, h.Source)
	}

	return &rtn, &(resp.ScrollId), nil
}

func extract_ids(re *regexp.Regexp, records *[]Record) []string {
	var rtn []string

	for _, rec := range *records {
		match := re.FindStringSubmatch(rec.MessageDetail)
		if match != nil {
			rtn = append(rtn, match[1])
		} else {
			log.Printf("No match on %s", rec.MessageDetail)
		}
	}
	return rtn
}

func main() {
	pageSize := 500
	startAt := 0

	vsUri := os.Getenv("VIDISPINE_URI")
	vsUser := os.Getenv("VIDISPINE_USER")
	vsPasswd := os.Getenv("VIDISPINE_PASSWD")
	indexName := os.Getenv("INDEX_NAME")
	environment := os.Getenv("ENVIRONMENT")

	if indexName == "" {
		log.Fatalf("Please set a Logstash index to query with the INDEX_NAME parameter")
	}

	if vsUri == "" || vsUser == "" || vsPasswd == "" {
		log.Printf("You should set VIDISPINE_URI, VIDISPINE_USER and VIDISPINE_PASSWD in order to get notification details from the server")
	}

	//set ELASTICSEARCH_URL to say where to connect to
	esclient, eserr := elasticsearch6.NewDefaultClient()

	if eserr != nil {
		log.Fatalf("Error creating client: %s", eserr)
	}

	esinfo, err := esclient.Info()
	if err != nil {
		log.Fatalf("Could not connect to cluster: %s", err)
	}

	if esinfo.IsError() {
		log.Fatalf("Error: %s", esinfo.String())
	}

	log.Printf("%s", esinfo)
	queryBuffer := make_query(environment)

	var atRecord = startAt
	var uniqueList []string
	var scrollId *string
	scrollId = nil

	vsComm := NewVSCommunicator(vsUri, vsUser, vsPasswd)

	for {
		records, newScrollId, _ := find_records(esclient, indexName, queryBuffer, scrollId, atRecord, pageSize)
		scrollId = newScrollId
		if len(*records) == 0 {
			break
		}

		re := regexp.MustCompile("notification=(\\w{2}-\\d+)")

		log.Printf("Got %d records: ", len(*records))

		id_list := extract_ids(re, records)
		tempList := append(uniqueList, id_list...)
		uniqueList = unique(tempList)
		log.Printf("Found %d ids: %s", len(uniqueList), uniqueList)
		atRecord += pageSize
	}

	if scrollId != nil {
		cleanup(esclient, scrollId)
	}

	for _, uniqId := range uniqueList {
		noti, err := vsComm.FindAndParseAnyNotification(uniqId)

		if err != nil {
			log.Fatalf("Could not retrieve information about %s from server: %s", uniqId, err)
		}

		if noti != nil {
			log.Printf("%s: %s", uniqId, noti.getInfoString())
		} else {
			log.Printf("%s: Not found", uniqId)
		}
	}

	log.Printf("Run completed")
}
