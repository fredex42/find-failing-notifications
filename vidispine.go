package main

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type VSCommunicator struct {
	URI      string
	User     string
	Password string
	client   *http.Client
}

type DownloadedResponse struct {
	BodyContent string
	StatusCode  int
}

/**
Initialise a new VSCommunicator object
*/
func NewVSCommunicator(uri string, user string, password string) *VSCommunicator {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	comm := VSCommunicator{uri, user, password, &http.Client{}}

	return &comm
}

/**
adds authentication headers to the request, sends it and downloads the response.
returns a DownloadedResponse object with the status code and downloaded body (as a string).
if error is set indicates a sending error, NOT an HTTP error (non-20x status)
*/
func (comm *VSCommunicator) authAndSend(req *http.Request) (*DownloadedResponse, error) {
	req.SetBasicAuth(comm.User, comm.Password)
	response, err := comm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resp := DownloadedResponse{}

	bodyContent, _ := ioutil.ReadAll(response.Body)
	resp.BodyContent = string(bodyContent)
	resp.StatusCode = response.StatusCode
	return &resp, nil
}

/**
tries to find a vidispine notification of the specific type.
returns a string of the xml NotificationDocument if one exists, nil if it does not, or an error.
on a 503/504 will sleep for 3s and retry.
*/
func (comm *VSCommunicator) FindSpecificNotification(notificationClass string, notificationId string) (*string, *string, error) {
	reqUri := comm.URI + "/API/" + notificationClass + "/notification/" + notificationId
	log.Printf("Requesting data from %s", reqUri)
	req, reqErr := http.NewRequest("GET", reqUri, nil)

	if reqErr != nil {
		return nil, nil, reqErr
	}

	dlResponse, sendErr := comm.authAndSend(req)
	if sendErr != nil {
		return nil, nil, sendErr
	}

	switch dlResponse.StatusCode {
	case 200:
		//log.Printf("Found %s notification for %s", notificationClass, notificationId)
		return &(dlResponse.BodyContent), &reqUri, nil
	case 404:
		//log.Printf("No %s notification found for id %s", notificationClass, notificationId)
		return nil, nil, nil
	case 500:
		log.Printf("Vidispine returned server error: %s", dlResponse.BodyContent)
		return nil, nil, errors.New("Vidispine Server error")
	case 503:
		log.Printf("Server is not available (503 error). Retrying after delay...")
		time.Sleep(time.Duration(3) * time.Second)
		return comm.FindSpecificNotification(notificationClass, notificationId)
	case 504:
		log.Printf("Server is not available (504 error). Retrying after delay...")
		time.Sleep(time.Duration(3) * time.Second)
		return comm.FindSpecificNotification(notificationClass, notificationId)
	default:
		log.Printf("Received unexpected status code %d", dlResponse.StatusCode)
		return nil, nil, errors.New("Unexpected status code")
	}
}

/**
looks over all notification classes to try to find a matching notification.
returns the first one that matches, or nil if nothing found.
*/
func (comm *VSCommunicator) FindAnyNotification(notificationId string) (*string, *string, error) {
	possibleClasses := []string{"item", "collection", "job", "storage", "storage/file", "quota", "group", "document", "deletion-lock"}

	for _, cls := range possibleClasses {
		xmldoc, notificationUrl, err := comm.FindSpecificNotification(cls, notificationId)
		if err != nil {
			return nil, nil, err
		}
		if xmldoc != nil {
			return xmldoc, notificationUrl, nil
		}
	}

	log.Printf("No notification of any type found for %s", notificationId)
	return nil, nil, nil
}

/**
same as FindAnyNotification but parses the returned string into a NotificationDocument object
*/
func (comm *VSCommunicator) FindAndParseAnyNotification(notificationId string) (*NotificationDocument, *string, error) {
	xmlString, notificationUrl, err := comm.FindAnyNotification(notificationId)

	if xmlString != nil {
		doc := NotificationDocument{}

		decodeErr := xml.Unmarshal([]byte(*xmlString), &doc)
		if decodeErr != nil {
			return nil, nil, decodeErr
		} else {
			return &doc, notificationUrl, err
		}
	} else {
		return nil, notificationUrl, err
	}
}
