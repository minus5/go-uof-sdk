package uof

// SnapshotComplete message indicates that all messages relating to an initiate
// request for odds from the RESTful API have been processed. The request_id
// parameter returned is the request_id that was specified in the
// initiate_request POST to the API. If no request_id parameter was specified in
// the original request, no request_id parameter is present. It is highly
// recommended to set a request_id to some number.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Snapshot+complete
type SnapshotComplete struct {
	Producer  Producer `xml:"product,attr" json:"producer"`
	Timestamp int      `xml:"timestamp,attr" json:"timestamp"`
	RequestID int      `xml:"request_id,attr" json:"requestID"`
}
