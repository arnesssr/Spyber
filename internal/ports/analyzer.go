// SPDX-License-Identifier: AGPL-3.0-only

package ports

type PageAnalysis struct {
	Emails           []string
	ContactLinks     []string
	CandidateLinks   []string
	EcommerceSignals []string
	EcommerceScore   int
	Text             string
}

type Analyzer interface {
	Analyze(baseURL string, body []byte) PageAnalysis
}
