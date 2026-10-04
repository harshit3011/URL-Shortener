package models

type Counter struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Sequence int64  `json:"sequence" bson:"sequence"`
}
