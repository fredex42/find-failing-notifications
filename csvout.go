package main

import (
	"encoding/csv"
	"log"
	"os"
)

type CsvData struct {
	doc             *NotificationDocument
	notificationUrl string
}

func WriteToCsv(filename string, docs *[]CsvData) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Could not create output file %s: %s", filename, err)
		return err
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, entry := range *docs {
		dataLine := []string{entry.notificationUrl, entry.doc.Action.Http.Url, entry.doc.Action.Http.Method, entry.doc.Action.Http.ContentType}
		err := writer.Write(dataLine)
		if err != nil {
			log.Printf("Could not write line to file: %s", err)
			return err
		}
	}
	return nil
}
