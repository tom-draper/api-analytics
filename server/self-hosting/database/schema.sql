CREATE TABLE IF NOT EXISTS users (
    api_key uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL,
    created_at timestamp with time zone,
    last_accessed timestamp with time zone
);

CREATE TABLE IF NOT EXISTS requests (
    request_id serial PRIMARY KEY,
    api_key uuid NOT NULL,
    method smallint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    path character varying(255) NOT NULL,
    status smallint NOT NULL,
    response_time smallint NOT NULL,
    framework smallint NOT NULL,
    hostname character varying(255),
    ip_address cidr,
    location character varying(2),
    user_id character varying(255),
    user_agent_id integer,
    user_hash character varying(32),
    referrer character varying(255)
);

CREATE TABLE IF NOT EXISTS user_agents (
    id serial PRIMARY KEY,
    user_agent character varying(255) UNIQUE
);

CREATE TABLE IF NOT EXISTS monitor (
    api_key uuid NOT NULL,
    url character varying(255) NOT NULL,
    secure boolean,
    ping boolean,
    created_at timestamp with time zone,
    PRIMARY KEY (api_key, url)
);

CREATE TABLE IF NOT EXISTS pings (
    api_key uuid NOT NULL,
    url character varying(255) NOT NULL,
    response_time integer,
    status smallint,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (api_key, url, created_at)
);

ALTER TABLE ONLY requests
    ADD CONSTRAINT requests_user_agent_id_fkey
    FOREIGN KEY (user_agent_id) REFERENCES user_agents(id);

CREATE INDEX IF NOT EXISTS api_key_index ON requests USING hash (api_key);
