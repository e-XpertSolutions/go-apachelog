package apachelog

import "time"

// An AccessLogEntry represents a line of an access log.
type AccessLogEntry struct {
	RemoteIPAddr        string            // Remote IP address
	LocalIPAddr         string            // Local IP address
	ResponseSize        int64             // Size of response in bytes, excluding HTTP headers
	Cookies             map[string]string // Cookies value
	ElapsedTime         int64             // Time taken to serve the request, in microseconds
	EnvVars             map[string]string // Content of the environment variables
	Filename            string            // Filename
	RemoteHost          string            // Remote host
	RequestProto        string            // Request protocol
	KeepAliveRequests   int64             // Number of keepalive requests handled on this connection
	RemoteLogname       string            // Remote logname. A "-" is returned, when not supplied
	RequestMethod       string            // Request method
	Port                string            // Canonical port of the server serving the request
	ProcessID           int               // Process ID of the child that serviced the request
	QueryString         string            // Query string (prepended with a ? if exists)
	Status              string            // Status
	Time                time.Time         // Time the request was received (standard english format)
	ElapsedTimeSec      int               // Time taken to serve the request, in seconds
	RemoteUser          string            // Remote user (from auth)
	URLPath             string            // URL path requested, not including any query string
	CanonicalServerName string            // Canonical ServerName of the server serving the request
	ServerName          string            // Server name according to the UseCanonicalName setting
	BytesReceived       int64             // Bytes received, including request and headers
	BytesSent           int64             // Bytes sent, including headers
}
