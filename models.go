package main

import "encoding/json"
import "fmt"

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
  ScrollId string `json:"_scroll_id"`
}

type HttpNotification struct {
  Synchronous bool `xml:"synchronous,attr"`
  Retry int `xml:"retry"`
  ContentType string `xml:"contentType"`
  Url string `xml:"url"`
  Method string `xml:"method"`
  Timeout int `xml:"timeout"`
}

type EJB struct {
  RawData string `xml:",innerxml"`
}

type JMS struct {
  RawData string `xml:",innerxml"`
}

type Javascript struct {
  RawData string `xml:",innerxml"`
}

type SQS struct {
  RawData string `xml:",innerxml"`
}

type Action struct {
  Http HttpNotification `xml:"http"`
  EJB EJB `xml:"ejb"`
  JMS JMS `xml:"jms"`
  Javascript Javascript `xml:"javascript"`
  SQS SQS `xml:"sqs"`
}

type ShapeTrigger struct {
  Modify *struct{} `xml:"modify"`

}

type JobTrigger struct {
  Update *struct{} `xml:"update"`
  Stop *struct{} `xml:"stop"`
  Finished *struct{} `xml:"finished"`
  Fail *struct{} `xml:"fail"`
  Create *struct{} `xml:"create"`
}

type Trigger struct {
  RawData string `xml:",innerxml"`
}

type NotificationDocument struct {
  Action Action `xml:"action"`
  Trigger Trigger `xml:"trigger"`
}

func (n *NotificationDocument) getInfoString ()(string) {
  if &(n.Action.Http)!=nil {
    return fmt.Sprintf("HTTP notification on %#v, call out to %s %s in %s format.", n.Trigger, n.Action.Http.Method, n.Action.Http.Url, n.Action.Http.ContentType)
  } else if &(n.Action.EJB)!=nil {
    return fmt.Sprintf("EJB notification on %#v", n.Trigger)
  } else if &(n.Action.JMS)!=nil {
    return fmt.Sprintf("JMS notification on %#v", n.Trigger)
  } else if &(n.Action.Javascript)!=nil {
    return fmt.Sprintf("Javascript notification on %#v", n.Trigger)
  } else if &(n.Action.SQS)!=nil {
    return fmt.Sprintf("SQS notification on %#v", n.Trigger)
  } else {
    return fmt.Sprintf("No action parameter on %#v", n.Trigger)
  }
}
