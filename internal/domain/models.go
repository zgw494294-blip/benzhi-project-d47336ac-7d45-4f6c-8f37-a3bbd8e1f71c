package domain

import (
	"encoding/json"
	"time"
)

type SampleBatch struct {
	BatchID           string              `json:"batchId"`
	Location          string              `json:"location"`
	CollectionWindow  string              `json:"collectionWindow"`
	Species           string              `json:"species"`
	SuspectedIssue    string              `json:"suspectedIssue"`
	Collector         string              `json:"collector"`
	Status            BatchStatus         `json:"status"`
	Version           int                 `json:"version"`
	CreatedAt         time.Time           `json:"createdAt"`
	UpdatedAt         time.Time           `json:"updatedAt"`
	Evidence          []FieldEvidence     `json:"evidence,omitempty"`
	Review            *ExpertReview       `json:"review,omitempty"`
	Certificate       *ReleaseCertificate `json:"certificate,omitempty"`
	OpenIssues        []string            `json:"openIssues,omitempty"`
	EvidenceScore     int                 `json:"evidenceScore"`
	EvidenceCheckedAt *time.Time          `json:"evidenceCheckedAt,omitempty"`
}

type FieldEvidence struct {
	EvidenceID   string             `json:"evidenceId"`
	BatchID      string             `json:"batchId"`
	SampleNumber string             `json:"sampleNumber"`
	Grid         string             `json:"grid"`
	PhotoDigest  string             `json:"photoDigest"`
	Environment  map[string]float64 `json:"environment"`
	Notes        string             `json:"notes"`
	SubmittedAt  time.Time          `json:"submittedAt"`
	Integrity    string             `json:"integrity"`
	Check        *EvidenceCheck     `json:"check,omitempty"`
}

type ExpertReview struct {
	ReviewID      string    `json:"reviewId"`
	BatchID       string    `json:"batchId"`
	Conclusion    string    `json:"conclusion"`
	RiskLevel     string    `json:"riskLevel"`
	Issues        []string  `json:"issues"`
	Rectification string    `json:"rectification"`
	ReviewOpinion string    `json:"reviewOpinion"`
	Reviewer      string    `json:"reviewer"`
	ReviewedAt    time.Time `json:"reviewedAt"`
}

type ReleaseCertificate struct {
	DispositionID   string          `json:"dispositionId"`
	BatchID         string          `json:"batchId"`
	Plan            string          `json:"plan"`
	Disposition     DispositionPlan `json:"disposition"`
	ExecutionWindow string          `json:"executionWindow"`
	FreezeDigest    string          `json:"freezeDigest"`
	FreezeVersion   string          `json:"freezeVersion"`
	Credential      string          `json:"credential"`
	Issuer          string          `json:"issuer"`
	IssuedAt        time.Time       `json:"issuedAt"`
}

type AuditEvent struct {
	EventID       string          `json:"eventId"`
	BatchID       string          `json:"batchId"`
	Sequence      int             `json:"sequence"`
	EventType     string          `json:"eventType"`
	Payload       any             `json:"payload"`
	PrevHash      string          `json:"prevHash"`
	Hash          string          `json:"hash"`
	OccurredAt    time.Time       `json:"occurredAt"`
	SchemaVersion int             `json:"schemaVersion"`
	HashPayload   json.RawMessage `json:"-"`
}
