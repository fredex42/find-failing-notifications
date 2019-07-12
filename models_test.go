package main

import (
  "encoding/xml"
  "testing"
)

func TestUnmarshalData(t *testing.T) {
  testData := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><NotificationDocument xmlns="http://xml.vidispine.com/schema/vidispine"><action><http synchronous="false"><retry>3</retry><contentType>application/json</contentType><url>http://10.235.51.233/gnm_businesslogic/shape/</url><method>POST</method><timeout>11</timeout></http></action><trigger><shape><modify></modify></shape></trigger></NotificationDocument>`

  doc := NotificationDocument{}

  err := xml.Unmarshal([]byte(testData), &doc)

  if err != nil {
    t.Fatalf("Could not unmarshal test data: %s", err)
  }

  if doc.Action.Http.Synchronous != false {
    t.Errorf("synchronous flag should be false")
  }

  if doc.Action.Http.Retry != 3 {
    t.Errorf("retry should be 3")
  }

  if doc.Action.Http.ContentType != "application/json" {
    t.Errorf("content type should be application/json")
  }

  if doc.Action.Http.Url != "http://10.235.51.233/gnm_businesslogic/shape/" {
    t.Errorf("url is incorrect")
  }

  if doc.Action.Http.Method != "POST" {
    t.Errorf("method is incorrect")
  }

  if doc.Action.Http.Timeout != 11 {
    t.Errorf("timeout is incorrect")
  }
}

func TestUnmarshalEJBData(t *testing.T) {
  testData := `<NotificationDocument xmlns="http://xml.vidispine.com/schema/vidispine">
   <action>
      <ejb synchronous="true">
         <bean>vidibrain.beans.MyBeanRemote</bean>
         <method>myMethod</method>
         <data>
            <key>param1</key>
            <value>value1</value>
         </data>
         <data>
            <key>param2</key>
            <value>value2</value>
         </data>
         <data>
            <key>param3</key>
            <value>value3</value>
         </data>
      </ejb>
   </action>
   <trigger>
      ...
   </trigger>
</NotificationDocument>`
  doc := NotificationDocument{}

  err := xml.Unmarshal([]byte(testData), &doc)
  if err != nil {
    t.Fatalf("Could not unmarshal test data: %s", err)
  }
}
