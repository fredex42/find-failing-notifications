package main

import "encoding/json"

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
