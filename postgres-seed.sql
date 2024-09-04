--
-- Table `jobs` contains scheduled jobs.
--
CREATE TABLE IF NOT EXISTS jobs (
    id       SERIAL PRIMARY KEY,
    checksum VARCHAR(64)    NOT NULL,
    job_type VARCHAR(16)    NOT NULL,
    label    TEXT           NOT NULL,
    last_run TIMESTAMP,
    metadata JSONB          NOT NULL,
    next_run TIMESTAMP,
    state    VARCHAR(16)    NOT NULL
);

ALTER TABLE jobs ADD CONSTRAINT jobs_unique_checksum UNIQUE (checksum);

CREATE INDEX jobs_type_idx
    ON jobs USING HASH (job_type);

--
-- Table `jobs_events` contains jobs' audit logs.
--
CREATE TABLE IF NOT EXISTS jobs_events (
    id        SERIAL PRIMARY KEY,
    event_msg TEXT          NOT NULL,
    job_id    INTEGER       NOT NULL REFERENCES jobs ON DELETE CASCADE,
    ts        TIMESTAMP     NOT NULL
);

--
-- Table `user_followers` contains connections who follow `account_id`.
--
CREATE TABLE IF NOT EXISTS user_followers (
    account_id BIGINT       NOT NULL,
    first_seen TIMESTAMP    NOT NULL,
    handler    TEXT         NOT NULL,
    last_seen  TIMESTAMP    NOT NULL,
    pic_url    TEXT,
    user_id    BIGINT       NOT NULL,

    PRIMARY KEY (account_id, user_id)
);

--
-- Table `user_following` contains connections followed by `account_id`.
--
CREATE TABLE IF NOT EXISTS user_following (
    account_id BIGINT       NOT NULL,
    first_seen TIMESTAMP    NOT NULL,
    handler    TEXT         NOT NULL,
    last_seen  TIMESTAMP    NOT NULL,
    pic_url    TEXT,
    user_id    BIGINT       NOT NULL,

    PRIMARY KEY (account_id, user_id)
);