package tracker

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

type Tracker interface {
	Track(*http.Request, *TrackMessage)
}

type TrackMessage struct {
	Path         string
	ResponseCode int

	Request *TrackHTTPRequest
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
