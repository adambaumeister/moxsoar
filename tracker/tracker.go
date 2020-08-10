package tracker

import (
	"github.com/adambaumeister/moxsoar/settings"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Tracker interface {
	Track(*http.Request, *TrackMessage)
}

type TrackMessage struct {
	Timestamp       int64
	TimestampString string
	Path            string
	ResponseCode    int

	Request *TrackHTTPRequest
}

func GetTracker(settings *settings.Settings) Tracker {
	/*
		Grab the available tracker. Attempts to connect to Elasticsearch and if that fails,
		returns the standard debug tracker.
	*/
	et, err := GetElkTracker(settings)

	if err == nil {
		return et
	}

	dt := GetDebugTracker()

	return dt
}

// Stripped down http.Request message to be in the right format
type TrackHTTPRequest struct {
	Method     string
	Form       url.Values
	Header     http.Header
	Body       string
	Host       string
	RemoteAddr string
	RequestURI string
}

func BuildTrackMessage(r *http.Request, tm *TrackMessage) *TrackMessage {
	b, _ := ioutil.ReadAll(r.Body)

	thr := TrackHTTPRequest{
		Method:     r.Method,
		Form:       r.Form,
		Header:     r.Header,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		RequestURI: r.RequestURI,
		Body:       string(b),
	}
	tm.Request = &thr

	return tm
}
