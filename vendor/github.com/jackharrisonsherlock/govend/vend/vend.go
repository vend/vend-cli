// Package vend handles interactions with the Vend API.
package vend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Client contains API authentication details.
type Client struct {
	Token        string
	DomainPrefix string
	TimeZone     string
}

// NewClient is called to pass authentication details to the manager.
func NewClient(Token, DomainPrefix, tz string) Client {
	return Client{Token, DomainPrefix, tz}
}

// NewRequest performs a request to a Vend API endpoint.
func (c *Client) NewRequest(method, url string, body interface{}) (*http.Request, error) {

	// Convert body into JSON
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bb := bytes.NewReader(b)

	req, err := http.NewRequest(method, url, bb)
	if err != nil {
		fmt.Printf("\nError creating http request: %s", err)
		return nil, err
	}

	// Request Headers
	req.Header.Set("User-Agent", "Vend CLI")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	return req, nil
}

// Do request
func (c *Client) Do(req *http.Request) ([]byte, error) {

	client := http.DefaultClient
	var attempt int
	var resp *http.Response
	var err error
	for {
		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("\nError performing request: %s", err)
			// Delays between attempts will be exponentially longer each time.
			attempt++
			delay := BackoffDuration(attempt)
			time.Sleep(delay)
		} else {
			break
		}
	}

	defer resp.Body.Close()
	ResponseCheck(resp.StatusCode)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("\nError while reading response body: %s\n", err)
		return nil, err
	}

	return responseBody, err
}

func (c Client) MakeRequest(method, url string, body interface{}) ([]byte, error) {
	req, err := c.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ResourcePage gets a single page of data from a 2.0 API resource using a version attribute.
func (c *Client) ResourcePage(version int64, method, resource string) ([]byte, int64, error) {

	url := c.urlFactory(version, "", resource)
	body, err := c.MakeRequest(method, url, nil)
	response := Payload{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshalling payload: %s", err)
		return nil, 0, err
	}

	data := response.Data
	version = response.Version["max"]

	return data, version, err
}

// ResourcePageFlake gets a single page of data from a 2.0 API resource using a Flake ID attribute.
func (c *Client) ResourcePageFlake(id, method, resource string) ([]byte, string, error) {

	// Build the URL for the resource page.
	url := c.urlFactoryFlake(id, resource)
	body, err := c.MakeRequest(method, url, nil)
	payload := map[string][]interface{}{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		fmt.Printf("\nError unmarshalling payload: %s", err)
		return nil, "", err
	}

	items := payload["data"]

	// Retrieve the last ID from the payload to be used to request subsequent page
	// **TODO** Last ID will be stripped as its included in the previous payload, need a better way to handle this
	i := items[(len(items) - 1)]
	m := i.(map[string]interface{})
	lastID := m["id"].(string)

	return body, lastID, err
}

// ResponseCheck checks the HTTP status codes of responses.
func ResponseCheck(statusCode int) bool {
	switch statusCode {
	case 200, 201:
		return true
	case 401:
		fmt.Printf("\nAccess denied - check personal API token. Status: %d", statusCode)
		os.Exit(0)
	case 404:
		fmt.Printf("\nURL not found - Status: %d", statusCode)
		os.Exit(0)
	case 429:
		fmt.Printf("\nRate limited by the Vend API :S Status: %d", statusCode)
	case 500:
		fmt.Printf("\nServer error. Status: %d", statusCode)
	case 502:
		fmt.Printf("\nServer received an invalid response :S Status: %d", statusCode)
		os.Exit(0)
	default:
		fmt.Printf("\nGot an unknown status code - Google it. Status: %d", statusCode)
	}
	return false
}

// BackoffDuration ...
func BackoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	seconds := math.Pow(float64(attempt), 3.5) + 5
	return time.Second * time.Duration(seconds)
}

// urlFactory creates a Vend API 2.0 URL based on a resource.
func (c *Client) urlFactory(version int64, objectID, resource string) string {
	// Page size is capped at ten thousand for all endpoints except sales which it is capped at five hundred.
	const (
		pageSize = 10000
		deleted  = true
	)

	// Using 2.x Endpoint.
	address := fmt.Sprintf("https://%s.vendhq.com/api/2.0/", c.DomainPrefix)
	query := url.Values{}
	query.Add("after", fmt.Sprintf("%d", version))

	if objectID != "" {
		address += fmt.Sprintf("%s/%s/products?%s", resource, objectID, query.Encode())
	} else {
		address += fmt.Sprintf("%s?%s", resource, query.Encode())
	}

	return address
}

// urlFactoryFlake creates a Vend API 2.0 URL based on a resource.
func (c *Client) urlFactoryFlake(id, resource string) string {
	// Page size is capped at ten thousand for all endpoints except sales which it is capped at five hundred.
	const (
		pageSize = 10000
		deleted  = true
	)

	// Using 2.x Endpoint.
	address := fmt.Sprintf("https://%s.vendhq.com/api/2.0/%s", c.DomainPrefix, resource)

	// Iterate through pages using the ?before= FLAKE ID attribute.
	if id != "" {
		query := url.Values{}
		query.Add("before", fmt.Sprintf("%s", id))
		address += fmt.Sprintf("?%s", query.Encode())
	}

	return address
}

// ImageUploadURLFactory creates the Vend URL for uploading an image.
func (c Client) ImageUploadURLFactory(productID string) string {
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/%s/actions/image_upload",
		c.DomainPrefix, productID)
	return url
}

// ParseVendDT converts the default Vend timestamp string into a go Time.time value.
func ParseVendDT(dt, tz string) time.Time {

	// Load store's timezone as location.
	loc, err := time.LoadLocation(tz)
	if err != nil {
		fmt.Printf("Error loading timezone as location: %s", err)
	}

	// Default Vend timedate layout.
	const longForm = "2006-01-02T15:04:05Z07:00"
	t, err := time.Parse(longForm, dt)
	if err != nil {
		log.Fatalf("Error parsing time into deafult timestamp: %s", err)
	}

	// Time in retailer's timezone.
	dtWithTimezone := t.In(loc)

	return dtWithTimezone

	// Time string with timezone removed.
	// timeStr := timeLoc.String()[0:19]
}
