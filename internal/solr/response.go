package solr

import "encoding/json"

type Response struct {
	Header   ResponseHeader `json:"responseHeader"`
	Response ResponseData   `json:"response"`
}

type ResponseHeader struct {
	Status int            `json:"status"`
	QTime  int            `json:"qTime"`
	Params ResponseParams `json:"params"`
}

type ResponseParams struct {
	JSON string `json:"json"`
	XML  string `json:"xml"`
}

type ResponseData struct {
	NumFound      int             `json:"numFound"`
	Start         int             `json:"start"`
	NumFoundExact bool            `json:"numFoundExact"`
	Docs          json.RawMessage `json:"docs"`
}
