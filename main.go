package main

import (
  elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
  "encoding/json"
  "log"
  "bytes"
  "context"
  "io"
  "regexp"
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
func make_query() (bytes.Buffer) {
  var buf bytes.Buffer

  query := map[string]interface{}{
    "query": map[string]interface{}{
      "bool": map[string]interface{}{
        "must": map[string]interface{}{
          "match": map[string]interface{}{
            "message_detail": "NETWORK_FAILURE",
          },
        },
        "filter": map[string]interface{}{
          "term": map[string]interface{}{
            "fields.type":"vidispine",
          },
        },
      },
    },
  }

  encodeErr := json.NewEncoder(&buf).Encode(query)
  if encodeErr != nil {
    log.Fatalf("Could not encode query: ", encodeErr)
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

func find_records(esclient *elasticsearch6.Client, indexName string, queryBuffer bytes.Buffer, offset int, limit int) (*[]Record, error) {
  result, err := esclient.Search(
    esclient.Search.WithContext(context.Background()),
    esclient.Search.WithIndex(indexName),
    esclient.Search.WithBody(&queryBuffer),
    esclient.Search.WithTrackTotalHits(true),
    esclient.Search.WithFrom(offset),
    esclient.Search.WithSize(limit),
  )

  if err != nil {
    return nil, err
  }
  defer result.Body.Close()

  if result.IsError() {
    e, err := generic_decode(result.Body)
    if err != nil {
      log.Fatalf("Error parsing response body: %s", err)
    } else {
      log.Fatalf("ES reported error: %s", e)
    }
  }

  var resp Response
  decodeErr := json.NewDecoder(result.Body).Decode(&resp)
  if decodeErr != nil {
    log.Fatalf("Error parsing response body: %s", decodeErr)
  }

  log.Printf("Got %d results: ", resp.Hits.Total)
  var rtn []Record
  for _,h := range resp.Hits.Hits {
    rtn = append(rtn, h.Source)
  }

  return &rtn, nil
}

func extract_ids(re *regexp.Regexp, records *[]Record) ([]string) {
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
  pageSize:=5
  startAt:=0
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
  queryBuffer := make_query()

  var atRecord = startAt
  var uniqueList []string

  for {
    records, _ := find_records(esclient,"logstash-2019.07.11", queryBuffer, atRecord, pageSize)

    if(len(*records)==0){
      break
    }

    re := regexp.MustCompile("notification=(\\w{2}-\\d+)")

    log.Printf("Got %d records: ", len(*records))

    id_list := extract_ids(re, records)
    tempList := append(uniqueList, id_list...)
    uniqueList = unique(tempList)
    log.Printf("Found ids: %s", uniqueList)
    atRecord+=pageSize
  }
  // for _,rec := range *records {
  //   log.Printf("%s", rec)
  // }
}
