package tracker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/settings"
	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"log"
	"net/http"
	"time"
)

type ElkTracker struct {
	Address string

	SSLConfig tls.Config

	Client *elasticsearch.Client
}

const ES_INDEX = "moxsoar_tracker_idx"

func GetElkTracker(settings *settings.Settings) (*ElkTracker, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			settings.Address,
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
		},
	}

	t := map[string]map[string]map[string]string{
		"properties": {
			"Timestamp": {
				"type":   "date",
				"format": "epoch_second",
			},
		},
	}
	b, err := json.Marshal(t)
	mappingb := bytes.NewReader(b)
	if settings.Username != "" {
		cfg.Username = settings.Username
		cfg.Password = settings.Password
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	timeout := make(chan bool, 1)
	errchan := make(chan error)
	mchan := make(chan bool)
	go func() {
		// Check if running
		res, err := es.Info()

		if err != nil {
			errchan <- err
			return
		}
		if res.IsError() {
			errchan <- fmt.Errorf("%v", res.String())
		}
		ib := map[string]string{}
		b, err := json.Marshal(ib)
		indexb := bytes.NewReader(b)
		// Create the index if not created
		ir := esapi.IndexRequest{
			Index: ES_INDEX,
			Body:  indexb,
		}
		res, err = ir.Do(context.Background(), es)
		if err != nil {
			errchan <- err
			return
		}
		if res.IsError() {
			errchan <- fmt.Errorf("%v", res.String())
		}

		mr := esapi.IndicesPutMappingRequest{
			Index: []string{ES_INDEX},
			Body:  mappingb,
		}
		res, err = mr.Do(context.Background(), es)
		if err != nil {
			errchan <- err
			return
		}

		if res.IsError() {
			errchan <- fmt.Errorf("%v", res.String())
		}

		mchan <- true
	}()

	go func() {
		time.Sleep(1 * time.Second)

		timeout <- true
	}()

	select {
	// If it works
	case <-mchan:
		fmt.Printf("Elasticsearch Server Connected!\n")
		et := ElkTracker{
			Client: es,
		}
		return &et, nil
	case e := <-errchan:
		fmt.Printf("Error connecting to elasticsearch: %v\n", e)
		return nil, e
	case <-timeout:
		fmt.Printf("Timeout connecting to elasticsearch!\n")
		return nil, fmt.Errorf("Failed to connect to ", settings.Address)
	}
}

func (t *ElkTracker) Track(r *http.Request, message *TrackMessage) {
	message = BuildTrackMessage(r, message)
	now := time.Now()
	sec := now.Unix()
	message.Timestamp = sec

	b, err := json.Marshal(message)
	br := bytes.NewReader(b)

	req := esapi.IndexRequest{
		Index: ES_INDEX,
		Body:  br,
	}

	res, err := req.Do(context.Background(), t.Client)
	if err != nil {
		log.Fatalf("Error getting response from Elasticsearch: %s", err)
	}
	if res.IsError() {
		fmt.Printf("Error adding document: %v", res.String())
	}
	defer res.Body.Close()
}
