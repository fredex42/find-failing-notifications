package main

import (
  elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
  "encoding/json"
  "log"
  "bytes"
  "context"
)

type Record struct  {
  ClassName string `json:"class"`
  Hostname string `json:"host.name"`
  Level string `json:"level"`
  Message string `json:"message"`
  MessageDetail string `json:"message_detail"`
}

type Hit struct {
  Index string `json:"_index"`
  Type string `json:"_type"`
  Id string `json:"_id"`
  Score float64 `json:"_score"`
  Source Record `json:"_source"`
}

type Hits struct {
  Total int `json:"total"`
  MaxScore float64 `json:"max_score"`
  Hits []Hit `json:"hits"`
}

type Response struct {
  Took int `json:"took"`
  TimedOut bool `json:"timed_out"`
  Shards json.RawMessage `json:"_shards"`
  Hits Hits `json:"hits"`
}


/*
{
	"query": {
	"bool": {
		"must": [
			{
				"match": {"message_detail": "NETWORK_FAILURE"}
			}
		],
		"filter": {
			"term": { "fields.type": "vidispine" }
		}
	}
}
}
*/
func find_records(esclient *elasticsearch6.Client, indexName string) (*[]Record, error) {
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

  result, err := esclient.Search(
    esclient.Search.WithContext(context.Background()),
    esclient.Search.WithIndex(indexName),
    esclient.Search.WithBody(&buf),
    esclient.Search.WithTrackTotalHits(true),
  )
  if err != nil {
    return nil, err
  }
  defer result.Body.Close()

  if result.IsError() {
    var e map[string]interface{}

    err := json.NewDecoder(result.Body).Decode(&e)
    if err != nil {
      log.Fatalf("Error parsing response body: %s", err)
    } else {
      log.Fatalf("ES reported error: %s", e)
    }
  }

  //log.Printf("%s", result)
  var resp Response
  decodeErr := json.NewDecoder(result.Body).Decode(&resp)
  if decodeErr != nil {
    log.Fatalf("Error parsing response body: %s", decodeErr)
  }

  log.Printf("Got %d results: ", resp.Hits.Total)
  var rtn []Record
  for _,h := range resp.Hits.Hits {
    //log.Printf("\t%s", h.Source)
    rtn = append(rtn, h.Source)
  }

  return &rtn, nil
}

func main() {

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

  records, err := find_records(esclient,"logstash-2019.07.11")

  log.Printf("Got %d records: ", len(*records))
  for _,rec := range *records {
    log.Printf("%s", rec)
  }
}
