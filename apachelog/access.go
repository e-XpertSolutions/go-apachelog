package apachelog

import (
	"net/url"
	"strings"
	"time"
)

// An AccessLogEntry represents a line of an access log.
type AccessLogEntry struct {
	RemoteIPAddr        string            // Remote IP address
	LocalIPAddr         string            // Local IP address
	ResponseSize        int64             // Size of response in bytes, excluding HTTP headers
	Cookies             map[string]string // Cookies value
	ElapsedTime         int64             // Time taken to serve the request, in microseconds
	EnvVars             map[string]string // Content of the environment variables
	Headers             map[string]string // Content of the request headers
	Filename            string            // Filename
	RemoteHost          string            // Remote host
	RequestProto        string            // Request protocol
	KeepAliveRequests   int64             // Number of keepalive requests handled on this connection
	RemoteLogname       string            // Remote logname. A "-" is returned, when not supplied
	RequestMethod       string            // Request method
	Port                string            // Canonical port of the server serving the request
	ProcessID           int64             // Process ID of the child that serviced the request
	QueryString         string            // Query string (prepended with a ? if exists)
	RequestFirstLine    RequestFirstLine  // First line of the request
	Status              string            // Status
	Time                time.Time         // Time the request was received (standard english format)
	ElapsedTimeSec      int64             // Time taken to serve the request, in seconds
	RemoteUser          string            // Remote user (from auth)
	URLPath             string            // URL path requested, not including any query string
	CanonicalServerName string            // Canonical ServerName of the server serving the request
	ServerName          string            // Server name according to the UseCanonicalName setting
	BytesReceived       int64             // Bytes received, including request and headers
	BytesSent           int64             // Bytes sent, including headers
}

// RequestFirstLine is a handy structure to hold the first line of the HTTP
// request. It provides accessors to access the HTTP method, the path and the
// protocol information contained in the raw first line.
type RequestFirstLine struct {
	raw string

	// Extracted fields
	method     string
	path       string
	pathParsed string
	proto      string

	parsedPath bool
	parsed     bool
}

// NewRequestFirstLine creates a new RequestFirstLine with the supplied raw
// first line.
//
// This function does not parse or do anything else than initializing the
// structure.
func NewRequestFirstLine(raw string) RequestFirstLine {
	return RequestFirstLine{
		raw: raw,
	}
}

// Method returns the method held in the HTTP request first line.
func (rfl *RequestFirstLine) Method() string {
	rfl.parse()
	return rfl.method
}

// RawPath returns the raw path held in the HTTP request first line.
func (rfl *RequestFirstLine) RawPath() string {
	rfl.parse()
	return rfl.path
}

// Path returns the parsed path (without url "percent encoding") held in the HTTP request first line.
func (rfl *RequestFirstLine) Path() string {
	rfl.parse()
	if rfl.parsedPath {
		return rfl.pathParsed
	}

	parsed, err := url.PathUnescape(rfl.path)
	if err != nil {
		rfl.pathParsed = rfl.path
	} else {
		rfl.pathParsed = parsed
	}
	rfl.parsedPath = true
	return rfl.pathParsed
}

// Protocol returns the protocol held in the HTTP request first line.
func (rfl *RequestFirstLine) Protocol() string {
	rfl.parse()
	return rfl.proto
}

func (rfl *RequestFirstLine) parse() {
	if rfl.parsed || rfl.raw == "" {
		return
	}

	s := strings.Split(rfl.raw, " ")
	if len(s) != 3 {
		return
	}
	rfl.method = s[0]
	rfl.path = s[1]
	rfl.proto = s[2]

	rfl.parsed = true
}

func (rfl RequestFirstLine) String() string {
	return rfl.raw
}
