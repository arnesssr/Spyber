// SPDX-License-Identifier: AGPL-3.0-only

package domain

type CompanyStatus string

const (
	CompanyCandidate CompanyStatus = "candidate"
	CompanyCrawled   CompanyStatus = "crawled"
	CompanyReview    CompanyStatus = "review"
	CompanyApproved  CompanyStatus = "approved"
	CompanyRejected  CompanyStatus = "rejected"
)

type ContactStatus string

const (
	ContactNeedsReview ContactStatus = "needs_review"
	ContactApproved    ContactStatus = "approved"
	ContactRejected    ContactStatus = "rejected"
	ContactSuppressed  ContactStatus = "suppressed"
)

type ContactType string

const (
	ContactGeneric ContactType = "generic"
	ContactNamed   ContactType = "named"
	ContactUnknown ContactType = "unknown"
)

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
)

type SourceStatus string

const (
	SourceActive   SourceStatus = "active"
	SourceDisabled SourceStatus = "disabled"
)
