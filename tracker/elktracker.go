package tracker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/adambaumeister/moxsoar/settings"
	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"log"
	"net/http"
)

type ElkTracker struct {
	Address   string
	SSLConfig tls.Config

	Client *elasticsearch.Client
}

func GetElkTracker(settings *settings.Settings) (*ElkTracker, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			settings.Address,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	res, err := es.Info()
	if err != nil {
		return nil, err
	}

	et := ElkTracker{
		Client: es,
	}
	defer res.Body.Close()
	return &et, nil
}

func (t *ElkTracker) Track(r *http.Request, message *TrackMessage) {
	message = BuildTrackMessage(r, message)

	b, err := json.Marshal(message)
	br := bytes.NewReader(b)
	req := esapi.IndexRequest{
		Index: "moxsoar_tracker_idx",
		Body:  br,
	}

	res, err := req.Do(context.Background(), t.Client)
	if err != nil {
		log.Fatalf("Error getting response from Elasticsearch: %s", err)
	}
	defer res.Body.Close()
}
