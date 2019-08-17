package model

//TrafficRequest structure
type TrafficRequest struct {
	EnvironmentID      int                     `json:"environmentId"`
	Domain             string                  `json:"domain"`
	ServiceName        string                  `json:"serviceName"`
	ContextPath        string                  `json:"contextPath"`
	HeaderName         string                  `json:"headerName"`
	HeaderValue        string                  `json:"headerValue"`
	HeaderReleaseName  string                  `json:"headerReleaseName"`
	DefaultReleaseName string                  `json:"defaultReleaseName"`
	Releases           []TrafficReleaseRequest `json:"releases"`
}

//TrafficReleaseRequest Structure
type TrafficReleaseRequest struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}
