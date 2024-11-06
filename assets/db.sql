CREATE TABLE IF NOT EXISTS virtual_devices (
    id VARCHAR(100) PRIMARY KEY,
    version INT NOT NULL,
    data jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS commits (
    seq_id BIGSERIAL PRIMARY KEY,
    commit_id BIGINT NULL,
    agg_id VARCHAR(100) NOT NULL,
    agg_type VARCHAR(100) NOT NULL,
    agg_version BIGINT NOT NULL,
    metadata JSONB NULL,
    events JSONB NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ix_commits_commit_id ON commits (commit_id); 
CREATE UNIQUE INDEX IF NOT EXISTS ix_commits_agg_type__agg_id__agg_version ON commits (agg_type, agg_id, agg_version); 

CREATE TABLE IF NOT EXISTS checkpoints (
    consumer_id VARCHAR(50) PRIMARY KEY,
    commit_id BIGINT NULL
);

CREATE TABLE IF NOT EXISTS controllers (
    id VARCHAR(100) PRIMARY KEY,
    lock_id SERIAL NOT NULL
);