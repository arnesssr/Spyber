CREATE INDEX idx_sources_country ON sources(country_code);
CREATE INDEX idx_companies_country_status ON companies(country_code, status);
CREATE INDEX idx_contacts_status_type ON contacts(status, contact_type);
CREATE INDEX idx_contacts_email ON contacts(email);
CREATE INDEX idx_evidence_company_type ON evidence(company_id, evidence_type);
CREATE INDEX idx_crawl_jobs_company_status ON crawl_jobs(company_id, status);
CREATE INDEX idx_audit_events_target ON audit_events(target_type, target_id);
