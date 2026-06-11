// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS sources (
		id TEXT PRIMARY KEY,
		country_code TEXT NOT NULL,
		source_type TEXT NOT NULL,
		url TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		UNIQUE (country_code, url)
	)`,
	`CREATE TABLE IF NOT EXISTS companies (
		id TEXT PRIMARY KEY,
		country_code TEXT NOT NULL,
		name TEXT NOT NULL,
		website_url TEXT NOT NULL,
		normalized_host TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL,
		ecommerce_score INTEGER NOT NULL DEFAULT 0,
		country_confidence INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS crawl_jobs (
		id TEXT PRIMARY KEY,
		company_id TEXT NOT NULL REFERENCES companies(id),
		url TEXT NOT NULL,
		status TEXT NOT NULL,
		failure_reason TEXT NOT NULL DEFAULT '',
		started_at TIMESTAMPTZ,
		finished_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS find_jobs (
		id TEXT PRIMARY KEY,
		country_code TEXT NOT NULL,
		sector TEXT NOT NULL,
		segment TEXT NOT NULL,
		query TEXT NOT NULL DEFAULT '',
		limit_count INTEGER NOT NULL,
		status TEXT NOT NULL,
		profile_key TEXT NOT NULL DEFAULT '',
		candidates INTEGER NOT NULL DEFAULT 0,
		created INTEGER NOT NULL DEFAULT 0,
		duplicates INTEGER NOT NULL DEFAULT 0,
		matched INTEGER NOT NULL DEFAULT 0,
		rejected INTEGER NOT NULL DEFAULT 0,
		fetched INTEGER NOT NULL DEFAULT 0,
		contacts INTEGER NOT NULL DEFAULT 0,
		direct_emails INTEGER NOT NULL DEFAULT 0,
		verified INTEGER NOT NULL DEFAULT 0,
		failures INTEGER NOT NULL DEFAULT 0,
		failure_reason TEXT NOT NULL DEFAULT '',
		started_at TIMESTAMPTZ,
		finished_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS fetch_tasks (
		id TEXT PRIMARY KEY,
		find_job_id TEXT NOT NULL REFERENCES find_jobs(id),
		company_id TEXT NOT NULL REFERENCES companies(id),
		url TEXT NOT NULL,
		purpose TEXT NOT NULL,
		status TEXT NOT NULL,
		attempts INTEGER NOT NULL DEFAULT 0,
		status_code INTEGER NOT NULL DEFAULT 0,
		bytes INTEGER NOT NULL DEFAULT 0,
		email_count INTEGER NOT NULL DEFAULT 0,
		link_count INTEGER NOT NULL DEFAULT 0,
		failure_reason TEXT NOT NULL DEFAULT '',
		started_at TIMESTAMPTZ,
		finished_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS contacts (
		id TEXT PRIMARY KEY,
		company_id TEXT NOT NULL REFERENCES companies(id),
		email TEXT NOT NULL,
		contact_type TEXT NOT NULL,
		status TEXT NOT NULL,
		source_url TEXT NOT NULL,
		first_seen_at TIMESTAMPTZ NOT NULL,
		last_seen_at TIMESTAMPTZ NOT NULL,
		UNIQUE (company_id, email)
	)`,
	`CREATE TABLE IF NOT EXISTS evidence (
		id TEXT PRIMARY KEY,
		company_id TEXT NOT NULL REFERENCES companies(id),
		evidence_type TEXT NOT NULL,
		value TEXT NOT NULL,
		source_url TEXT NOT NULL,
		confidence INTEGER NOT NULL,
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS suppression (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		reason TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS exports (
		id TEXT PRIMARY KEY,
		country_code TEXT NOT NULL,
		format TEXT NOT NULL,
		filters TEXT NOT NULL,
		row_count INTEGER NOT NULL,
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS audit_events (
		id TEXT PRIMARY KEY,
		actor TEXT NOT NULL,
		action TEXT NOT NULL,
		target_type TEXT NOT NULL,
		target_id TEXT NOT NULL,
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_sources_country ON sources(country_code)`,
	`CREATE INDEX IF NOT EXISTS idx_companies_country_status ON companies(country_code, status)`,
	`CREATE INDEX IF NOT EXISTS idx_contacts_status_type ON contacts(status, contact_type)`,
	`CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(email)`,
	`CREATE INDEX IF NOT EXISTS idx_evidence_company_type ON evidence(company_id, evidence_type)`,
	`CREATE INDEX IF NOT EXISTS idx_crawl_jobs_company_status ON crawl_jobs(company_id, status)`,
	`CREATE INDEX IF NOT EXISTS idx_find_jobs_country_status ON find_jobs(country_code, status)`,
	`CREATE INDEX IF NOT EXISTS idx_fetch_tasks_job_status ON fetch_tasks(find_job_id, status)`,
	`CREATE INDEX IF NOT EXISTS idx_fetch_tasks_company ON fetch_tasks(company_id)`,
	`CREATE INDEX IF NOT EXISTS idx_audit_events_target ON audit_events(target_type, target_id)`,
}
