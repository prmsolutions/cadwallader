package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic/v7"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type StatusResponse struct {
	Services []StatusBlob
}

type StatusBlob struct {
	ServiceName string         `json:"serviceName"`
	Status      string         `json:"status"`
	History     []StatusReport `json:"history"`
}

type StatusReport struct {
	Timestamp time.Time `json:"@timestamp"`
	Up        int64     `json:"up"`
	Down      int64     `json:"down"`
}

type Service struct {
	Name  string `yaml:"name"`
	Index string `yaml:"index"`
}

var configuration Config
var configPath string

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", printMessage)
	r.HandleFunc("/status", getStatusCheckData)
	r.Use(loggingMiddleware)

	configPtr := flag.String("c", "config.yml", "path to the configuration file")
 	flag.Parse()
 	configPath = *configPtr

 	config := configuration.loadConfig(configPath)

	serverAddress := fmt.Sprintf("%s:%s", config.Server.Hostname, config.Server.Port)
	fmt.Printf("Starting server @ %s\n", serverAddress)

	http.ListenAndServe(serverAddress, r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)

		log.Printf("%s %s %s %v %s", r.RemoteAddr, r.Method, r.RequestURI, r.Header.Get("User-Agent"), m.Duration)
		next.ServeHTTP(w, r)
	})
}

func createEsClient() *elastic.Client {
	config := configuration.loadConfig(configPath)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	var hostname strings.Builder
	hostname.WriteString("https://")
	hostname.WriteString(config.Elasticsearch.Hostname)
	hostname.WriteString(":")
	hostname.WriteString(config.Elasticsearch.Port)

	es, err := elastic.NewClient(
		elastic.SetHttpClient(client),
		elastic.SetURL(hostname.String()),
		elastic.SetScheme("https"),
		elastic.SetBasicAuth(os.Getenv("ELASTIC_USERNAME"), os.Getenv("ELASTIC_PASSWORD")),
		elastic.SetSniff(false),
	)

	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}

	return es
}

func getStatusCheckData(w http.ResponseWriter, r *http.Request) {
	config := configuration.loadConfig(configPath)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	response := StatusResponse{}

	for _, service := range config.Services {
		serviceBlob := computeServiceUptime(service.Name, service.Index)
		response.Services = append(response.Services, serviceBlob)
	}

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func computeServiceUptime(domain string, index string) StatusBlob {
	es := createEsClient()
	ctx := context.Background()

	matchQuery := elastic.NewTermQuery("url.domain", domain)
	rangeQuery := elastic.NewRangeQuery("@timestamp").Gt("now-7d")

	boolQuery := elastic.NewBoolQuery()
	boolQuery.Must(matchQuery, rangeQuery)

	statusAgg := elastic.NewTermsAggregation().Field("monitor.status")
	timeAgg := elastic.NewAutoDateHistogramAggregation().
		Field("@timestamp").
		SubAggregation("status", statusAgg)

	result, err := es.Search().
		Index(index).
		Query(boolQuery).
		Aggregation("time", timeAgg).
		Size(0).
		Pretty(true).
		Do(ctx)

	if err != nil {
		panic(err)
	}

	response := StatusBlob{}
	response.Status = "up"
	response.ServiceName = domain

	parseResult(result, &response)

	return response
}

func parseResult(result *elastic.SearchResult, response *StatusBlob) {
	raw := result.Aggregations["time"]
	var agg elastic.AggregationBucketKeyItems
	err := json.Unmarshal(raw, &agg)
	if err != nil {
		fmt.Printf("Unmarshal failed: %v\n", err)
		panic(err)
	}

	for _, item := range agg.Buckets {
		var subAgg elastic.AggregationBucketKeyItems
		err := json.Unmarshal(item.Aggregations["status"], &subAgg)
		if err != nil {
			fmt.Printf("Unmarshal failed: %v\n", err)
			panic(err)
		}

		update := StatusReport{}
		update.Timestamp = time.Unix(int64(item.Key.(float64)/1000.0), 0)

		for _, subItem := range subAgg.Buckets {
			if subItem.Key == "up" {
				update.Up = subItem.DocCount
			}
			if subItem.Key == "down" {
				update.Down = subItem.DocCount
			}
		}

		response.History = append(response.History, update)
	}
}

func printMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Cadwallader is running!")
}
