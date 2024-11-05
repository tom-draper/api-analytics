--
-- PostgreSQL database dump
--

-- Dumped from database version 14.11 (Ubuntu 14.11-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 14.11 (Ubuntu 14.11-0ubuntu0.22.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: monitor; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.monitor (
    api_key uuid NOT NULL,
    url character varying(255) NOT NULL,
    secure boolean,
    ping boolean,
    created_at timestamp with time zone
);


ALTER TABLE public.monitor OWNER TO postgres;

--
-- Name: pings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pings (
    api_key uuid NOT NULL,
    url character varying(255) NOT NULL,
    response_time integer,
    status smallint,
    created_at timestamp with time zone NOT NULL
);


ALTER TABLE public.pings OWNER TO postgres;

--
-- Name: requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.requests (
    request_id integer NOT NULL,
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
    user_agent_id integer
);


ALTER TABLE public.requests OWNER TO postgres;

--
-- Name: requests_request_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.requests_request_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.requests_request_id_seq OWNER TO postgres;

--
-- Name: requests_request_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.requests_request_id_seq OWNED BY public.requests.request_id;


--
-- Name: user_agents; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_agents (
    id integer NOT NULL,
    user_agent character varying(255)
);


ALTER TABLE public.user_agents OWNER TO postgres;

--
-- Name: user_agents_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_agents_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_agents_id_seq OWNER TO postgres;

--
-- Name: user_agents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_agents_id_seq OWNED BY public.user_agents.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    api_key uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone,
    last_accessed timestamp with time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: requests request_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.requests ALTER COLUMN request_id SET DEFAULT nextval('public.requests_request_id_seq'::regclass);


--
-- Name: user_agents id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_agents ALTER COLUMN id SET DEFAULT nextval('public.user_agents_id_seq'::regclass);


--
-- Name: monitor monitor_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.monitor
    ADD CONSTRAINT monitor_pkey PRIMARY KEY (api_key, url);


--
-- Name: pings pings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pings
    ADD CONSTRAINT pings_pkey PRIMARY KEY (api_key, url, created_at);


--
-- Name: requests requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.requests
    ADD CONSTRAINT requests_pkey PRIMARY KEY (request_id);


--
-- Name: user_agents user_agents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_agents
    ADD CONSTRAINT user_agents_pkey PRIMARY KEY (id);


--
-- Name: user_agents user_agents_user_agent_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_agents
    ADD CONSTRAINT user_agents_user_agent_key UNIQUE (user_agent);


--
-- Name: api_key_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX api_key_index ON public.requests USING hash (api_key);


--
-- Name: requests requests_user_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.requests
    ADD CONSTRAINT requests_user_agent_id_fkey FOREIGN KEY (user_agent_id) REFERENCES public.user_agents(id);


--
-- PostgreSQL database dump complete
--


