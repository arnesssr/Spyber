CREATE TABLE sources (
  id TEXT PRIMARY KEY,
  country_code TEXT NOT NULL,
  source_type TEXT NOT NULL,
  url TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (country_code, url)
);

CREATE TABLE companies (
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
);

CREATE TABLE crawl_jobs (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id),
  url TEXT NOT NULL,
  status TEXT NOT NULL,
  failure_reason TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE contacts (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id),
  email TEXT NOT NULL,
  contact_type TEXT NOT NULL,
  status TEXT NOT NULL,
  source_url TEXT NOT NULL,
  first_seen_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  UNIQUE (company_id, email)
);

CREATE TABLE evidence (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id),
  evidence_type TEXT NOT NULL,
  value TEXT NOT NULL,
  source_url TEXT NOT NULL,
  confidence INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE suppression (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  reason TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE exports (
  id TEXT PRIMARY KEY,
  country_code TEXT NOT NULL,
  format TEXT NOT NULL,
  filters TEXT NOT NULL,
  row_count INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE audit_events (
  id TEXT PRIMARY KEY,
  actor TEXT NOT NULL,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  metadata TEXT NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL
);
