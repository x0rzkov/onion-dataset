package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/urfave/cli/v2"

	tlog "github.com/creekorful/trandoshan/internal/log"
)

var (
	protocolRegex = regexp.MustCompile("https?://")
)

// Represent a resource in elasticsearch
type resourceIndex struct {
	URL   string    `json:"url"`
	Body  string    `json:"body"`
	Title string    `json:"title"`
	Time  time.Time `json:"time"`
}

// ResourceDto represent a resource as given by the API
type ResourceDto struct {
	URL   string    `json:"url"`
	Body  string    `json:"body"`
	Title string    `json:"title"`
	Time  time.Time `json:"time"`
}

// GetApp return the api app
func GetApp() *cli.App {
	return &cli.App{
		Name:    "trandoshan-api",
		Version: "0.0.1",
		Usage:   "", // TODO
		Flags: []cli.Flag{
			tlog.GetLogFlag(),
			&cli.StringFlag{
				Name:     "elasticsearch-uri",
				Usage:    "URI to the Elasticsearch server",
				Required: true,
			},
		},
		Action: execute,
	}
}

func execute(ctx *cli.Context) error {
	e := echo.New()
	e.HideBanner = true

	// Configure logger
	switch ctx.String("log-level") {
	case "debug":
		e.Logger.SetLevel(log.DEBUG)
	case "info":
		e.Logger.SetLevel(log.INFO)
	case "warn":
		e.Logger.SetLevel(log.WARN)
	case "error":
		e.Logger.SetLevel(log.ERROR)
	}

	e.Logger.Infof("Starting trandoshan-api v%s", ctx.App.Version)

	e.Logger.Debugf("Using elasticsearch server at: %s", ctx.String("elasticsearch-uri"))

	// Create Elasticsearch client
	es, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{ctx.String("elasticsearch-uri")}})
	if err != nil {
		e.Logger.Errorf("Error while creating elasticsearch client: %s", err)
		return err
	}

	// Add endpoints
	e.GET("/v1/resources", searchResources(es))
	e.POST("/v1/resources", addResource(es))

	e.Logger.Info("Successfully initialized trandoshan-api. Waiting for requests")

	return e.Start(":8080")
}

func searchResources(es *elasticsearch.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		url := c.QueryParam("url")

		var buf bytes.Buffer
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"url": url,
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			c.Logger().Errorf("Error encoding query: %s", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		// Perform the search request.
		res, err := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex("resources"),
			es.Search.WithBody(&buf),
		)
		if err != nil || res.IsError() {
			c.Logger().Errorf("Error getting response: %s", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			c.Logger().Errorf("Error parsing the response body: %s", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		var urls []ResourceDto
		for _, rawHit := range resp["hits"].(map[string]interface{})["hits"].([]interface{}) {
			rawSrc := rawHit.(map[string]interface{})["_source"].(map[string]interface{})

			res := ResourceDto{
				URL:   rawSrc["url"].(string),
				Body:  rawSrc["body"].(string),
				Title: rawSrc["title"].(string),
				Time:  time.Time{}, // TODO
			}

			urls = append(urls, res)
		}

		return c.JSON(http.StatusOK, urls)
	}
}

func addResource(es *elasticsearch.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var resourceDto ResourceDto
		if err := json.NewDecoder(c.Request().Body).Decode(&resourceDto); err != nil {
			c.Logger().Errorf("Error while un-marshaling resource: %s", err)
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		c.Logger().Debugf("Saving resource %s", resourceDto.URL)

		// TODO store on file system

		// Create Elasticsearch document
		doc := resourceIndex{
			URL:   protocolRegex.ReplaceAllLiteralString(resourceDto.URL, ""),
			Body:  resourceDto.Body,
			Title: extractTitle(resourceDto.Body),
			Time:  time.Now(),
		}

		// Serialize document into json
		docBytes, err := json.Marshal(&doc)
		if err != nil {
			c.Logger().Errorf("Error while serializing document into json: %s", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		// Use Elasticsearch to index document
		req := esapi.IndexRequest{
			Index:   "resources",
			Body:    bytes.NewReader(docBytes),
			Refresh: "true",
		}
		res, err := req.Do(context.Background(), es)
		if err != nil {
			c.Logger().Errorf("Error while creating elasticsearch index: %s", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		defer res.Body.Close()

		c.Logger().Debugf("Successfully saved resource %s", resourceDto.URL)

		return c.NoContent(http.StatusCreated)
	}
}

// extract title from html body
func extractTitle(body string) string {
	cleanBody := strings.ToLower(body)

	if strings.Index(cleanBody, "<title>") == -1 || strings.Index(cleanBody, "</title>") == -1 {
		return ""
	}

	// TODO improve
	startPos := strings.Index(cleanBody, "<title>") + len("<title>")
	endPos := strings.Index(cleanBody, "</title>")

	return body[startPos:endPos]
}
