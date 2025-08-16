--
-- PostgreSQL database dump
--

-- Dumped from database version 14.18 (Ubuntu 14.18-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 14.18 (Ubuntu 14.18-0ubuntu0.22.04.1)

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

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: update_user_news_preferences_updated_at(); Type: FUNCTION; Schema: public; Owner: most3mr
--

CREATE FUNCTION public.update_user_news_preferences_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_user_news_preferences_updated_at() OWNER TO most3mr;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: episode_tracking; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.episode_tracking (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    episode_id uuid,
    user_id uuid,
    watched boolean DEFAULT false NOT NULL,
    rating integer,
    notes text,
    watched_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT episode_tracking_rating_check CHECK (((rating >= 1) AND (rating <= 10)))
);


ALTER TABLE public.episode_tracking OWNER TO most3mr;

--
-- Name: episodes; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.episodes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    show_id uuid,
    user_id uuid,
    external_id integer NOT NULL,
    name text NOT NULL,
    season integer NOT NULL,
    number integer NOT NULL,
    summary text,
    airdate date,
    runtime integer,
    image_url text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    show_type character varying(20) DEFAULT 'tv'::character varying NOT NULL,
    filler boolean DEFAULT false NOT NULL,
    recap boolean DEFAULT false NOT NULL,
    CONSTRAINT episodes_type_check CHECK (((show_type)::text = ANY ((ARRAY['tv'::character varying, 'anime'::character varying])::text[])))
);


ALTER TABLE public.episodes OWNER TO most3mr;

--
-- Name: habits; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.habits (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    name text NOT NULL,
    scheduled_days jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.habits OWNER TO most3mr;

--
-- Name: habits_completions; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.habits_completions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    habit_id uuid,
    user_id uuid,
    completed boolean NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.habits_completions OWNER TO most3mr;

--
-- Name: market_data; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.market_data (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    symbol text NOT NULL,
    current_price numeric(15,6),
    change_amount numeric(15,6),
    change_percent numeric(8,4),
    volume bigint,
    market_cap bigint,
    last_updated timestamp without time zone DEFAULT now()
);


ALTER TABLE public.market_data OWNER TO most3mr;

--
-- Name: market_watchlist; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.market_watchlist (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    symbol text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT market_watchlist_type_check CHECK ((type = ANY (ARRAY['stock'::text, 'crypto'::text, 'forex'::text])))
);


ALTER TABLE public.market_watchlist OWNER TO most3mr;

--
-- Name: medication_logs; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.medication_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    medication_id uuid,
    user_id uuid,
    taken boolean NOT NULL,
    scheduled_time text,
    actual_time timestamp without time zone,
    date date NOT NULL,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.medication_logs OWNER TO most3mr;

--
-- Name: medications; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.medications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    name text NOT NULL,
    dosage text NOT NULL,
    scheduled_days jsonb NOT NULL,
    times_per_day integer DEFAULT 1 NOT NULL,
    time_intervals text[],
    duration_type text NOT NULL,
    start_date date NOT NULL,
    end_date date,
    notes text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT medications_duration_type_check CHECK ((duration_type = ANY (ARRAY['lifetime'::text, 'limited'::text])))
);


ALTER TABLE public.medications OWNER TO most3mr;

--
-- Name: mood_ratings; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.mood_ratings (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    rating integer NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT mood_ratings_rating_check CHECK (((rating >= 1) AND (rating <= 10)))
);


ALTER TABLE public.mood_ratings OWNER TO most3mr;

--
-- Name: news_articles; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.news_articles (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    source_id uuid,
    title text NOT NULL,
    content text,
    summary text,
    original_url text NOT NULL,
    image_url text,
    published_at timestamp without time zone,
    language text NOT NULL,
    category text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    thumbnail_url text,
    full_content text,
    author_name text,
    keywords text,
    date_modified timestamp with time zone,
    CONSTRAINT news_articles_language_check CHECK ((language = ANY (ARRAY['ar'::text, 'en'::text])))
);


ALTER TABLE public.news_articles OWNER TO most3mr;

--
-- Name: COLUMN news_articles.thumbnail_url; Type: COMMENT; Schema: public; Owner: most3mr
--

COMMENT ON COLUMN public.news_articles.thumbnail_url IS 'URL to article thumbnail image';


--
-- Name: COLUMN news_articles.full_content; Type: COMMENT; Schema: public; Owner: most3mr
--

COMMENT ON COLUMN public.news_articles.full_content IS 'Complete article body text from JSON-LD';


--
-- Name: COLUMN news_articles.author_name; Type: COMMENT; Schema: public; Owner: most3mr
--

COMMENT ON COLUMN public.news_articles.author_name IS 'Article author name from JSON-LD';


--
-- Name: COLUMN news_articles.keywords; Type: COMMENT; Schema: public; Owner: most3mr
--

COMMENT ON COLUMN public.news_articles.keywords IS 'Article keywords from JSON-LD';


--
-- Name: COLUMN news_articles.date_modified; Type: COMMENT; Schema: public; Owner: most3mr
--

COMMENT ON COLUMN public.news_articles.date_modified IS 'Last modification date from JSON-LD';


--
-- Name: news_sources; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.news_sources (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    url text NOT NULL,
    language text NOT NULL,
    category text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    fetch_frequency_hours integer DEFAULT 2 NOT NULL,
    last_fetched timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT news_sources_language_check CHECK ((language = ANY (ARRAY['ar'::text, 'en'::text]))),
    CONSTRAINT news_sources_type_check CHECK ((type = ANY (ARRAY['rss'::text, 'api'::text, 'scraper'::text])))
);


ALTER TABLE public.news_sources OWNER TO most3mr;

--
-- Name: notes; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.notes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    date date NOT NULL,
    text text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.notes OWNER TO most3mr;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    data jsonb,
    read boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.notifications OWNER TO most3mr;

--
-- Name: projects; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.projects (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.projects OWNER TO most3mr;

--
-- Name: shows; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.shows (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    external_id integer NOT NULL,
    name text NOT NULL,
    summary text,
    image_url text,
    status text,
    premiered date,
    ended date,
    network text,
    genres jsonb,
    rating jsonb,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    last_episode_sync timestamp without time zone,
    last_info_sync timestamp without time zone,
    show_type character varying(20) DEFAULT 'tv'::character varying NOT NULL,
    CONSTRAINT shows_type_check CHECK (((show_type)::text = ANY ((ARRAY['tv'::character varying, 'anime'::character varying])::text[])))
);


ALTER TABLE public.shows OWNER TO most3mr;

--
-- Name: sync_logs; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.sync_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    show_id uuid,
    sync_type character varying(20) NOT NULL,
    status character varying(20) NOT NULL,
    episodes_added integer DEFAULT 0,
    info_updated boolean DEFAULT false,
    error_message text,
    sync_duration_ms integer,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.sync_logs OWNER TO most3mr;

--
-- Name: sync_settings; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.sync_settings (
    id integer NOT NULL,
    setting_key character varying(50) NOT NULL,
    setting_value text NOT NULL,
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.sync_settings OWNER TO most3mr;

--
-- Name: sync_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: most3mr
--

CREATE SEQUENCE public.sync_settings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.sync_settings_id_seq OWNER TO most3mr;

--
-- Name: sync_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: most3mr
--

ALTER SEQUENCE public.sync_settings_id_seq OWNED BY public.sync_settings.id;


--
-- Name: task_attachments; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.task_attachments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    filename text NOT NULL,
    file_path text NOT NULL,
    file_size bigint NOT NULL,
    mime_type text NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.task_attachments OWNER TO most3mr;

--
-- Name: task_comments; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.task_comments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    comment text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.task_comments OWNER TO most3mr;

--
-- Name: task_dependencies; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.task_dependencies (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    depends_on_task_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT task_dependencies_no_self_reference CHECK ((task_id <> depends_on_task_id))
);


ALTER TABLE public.task_dependencies OWNER TO most3mr;

--
-- Name: tasks; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.tasks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    parent_task_id uuid,
    title text NOT NULL,
    description text,
    status text DEFAULT 'Not Started'::text NOT NULL,
    priority text DEFAULT 'None'::text NOT NULL,
    due_date timestamp without time zone,
    completed boolean DEFAULT false NOT NULL,
    display_order integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    collapsed boolean DEFAULT false NOT NULL,
    CONSTRAINT tasks_priority_check CHECK ((priority = ANY (ARRAY['None'::text, 'Low'::text, 'Medium'::text, 'High'::text]))),
    CONSTRAINT tasks_status_check CHECK ((status = ANY (ARRAY['Not Started'::text, 'In Progress'::text, 'Blocked'::text, 'In Review'::text, 'Completed'::text])))
);


ALTER TABLE public.tasks OWNER TO most3mr;

--
-- Name: todos; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.todos (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    text text NOT NULL,
    completed boolean NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.todos OWNER TO most3mr;

--
-- Name: user_interests; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.user_interests (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    interest_name text NOT NULL,
    keywords jsonb,
    sources jsonb,
    priority integer DEFAULT 1 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT user_interests_priority_check CHECK (((priority >= 1) AND (priority <= 5)))
);


ALTER TABLE public.user_interests OWNER TO most3mr;

--
-- Name: user_news_preferences; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.user_news_preferences (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    source_id uuid NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.user_news_preferences OWNER TO most3mr;

--
-- Name: users; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    password text NOT NULL,
    display_name text NOT NULL,
    avatar_url text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    email character varying(255) NOT NULL
);


ALTER TABLE public.users OWNER TO most3mr;

--
-- Name: workout_logs; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.workout_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    completed_exercises jsonb NOT NULL,
    cardio jsonb NOT NULL,
    weight double precision NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    name text DEFAULT ''::text NOT NULL
);


ALTER TABLE public.workout_logs OWNER TO most3mr;

--
-- Name: workouts; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.workouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    name text NOT NULL,
    day text NOT NULL,
    exercises jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    display_order integer DEFAULT 0
);


ALTER TABLE public.workouts OWNER TO most3mr;

--
-- Name: sync_settings id; Type: DEFAULT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.sync_settings ALTER COLUMN id SET DEFAULT nextval('public.sync_settings_id_seq'::regclass);


--
-- Name: episode_tracking episode_tracking_episode_user_unique; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_episode_user_unique UNIQUE (episode_id, user_id);


--
-- Name: episode_tracking episode_tracking_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_pkey PRIMARY KEY (id);


--
-- Name: episodes episodes_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_pkey PRIMARY KEY (id);


--
-- Name: habits_completions habits_completions_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits_completions
    ADD CONSTRAINT habits_completions_pkey PRIMARY KEY (id);


--
-- Name: habits habits_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits
    ADD CONSTRAINT habits_pkey PRIMARY KEY (id);


--
-- Name: market_data market_data_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.market_data
    ADD CONSTRAINT market_data_pkey PRIMARY KEY (id);


--
-- Name: market_data market_data_symbol_unique; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.market_data
    ADD CONSTRAINT market_data_symbol_unique UNIQUE (symbol);


--
-- Name: market_watchlist market_watchlist_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.market_watchlist
    ADD CONSTRAINT market_watchlist_pkey PRIMARY KEY (id);


--
-- Name: medication_logs medication_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_pkey PRIMARY KEY (id);


--
-- Name: medications medications_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medications
    ADD CONSTRAINT medications_pkey PRIMARY KEY (id);


--
-- Name: mood_ratings mood_ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.mood_ratings
    ADD CONSTRAINT mood_ratings_pkey PRIMARY KEY (id);


--
-- Name: news_articles news_articles_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.news_articles
    ADD CONSTRAINT news_articles_pkey PRIMARY KEY (id);


--
-- Name: news_articles news_articles_url_unique; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.news_articles
    ADD CONSTRAINT news_articles_url_unique UNIQUE (original_url);


--
-- Name: news_sources news_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.news_sources
    ADD CONSTRAINT news_sources_pkey PRIMARY KEY (id);


--
-- Name: notes notes_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: shows shows_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.shows
    ADD CONSTRAINT shows_pkey PRIMARY KEY (id);


--
-- Name: sync_logs sync_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.sync_logs
    ADD CONSTRAINT sync_logs_pkey PRIMARY KEY (id);


--
-- Name: sync_settings sync_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.sync_settings
    ADD CONSTRAINT sync_settings_pkey PRIMARY KEY (id);


--
-- Name: sync_settings sync_settings_setting_key_key; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.sync_settings
    ADD CONSTRAINT sync_settings_setting_key_key UNIQUE (setting_key);


--
-- Name: task_attachments task_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_pkey PRIMARY KEY (id);


--
-- Name: task_comments task_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_pkey PRIMARY KEY (id);


--
-- Name: task_dependencies task_dependencies_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_pkey PRIMARY KEY (id);


--
-- Name: task_dependencies task_dependencies_unique_pair; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_unique_pair UNIQUE (task_id, depends_on_task_id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: todos todos_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_pkey PRIMARY KEY (id);


--
-- Name: user_interests user_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.user_interests
    ADD CONSTRAINT user_interests_pkey PRIMARY KEY (id);


--
-- Name: user_news_preferences user_news_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.user_news_preferences
    ADD CONSTRAINT user_news_preferences_pkey PRIMARY KEY (id);


--
-- Name: user_news_preferences user_news_preferences_user_id_source_id_key; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.user_news_preferences
    ADD CONSTRAINT user_news_preferences_user_id_source_id_key UNIQUE (user_id, source_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: workout_logs workout_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workout_logs
    ADD CONSTRAINT workout_logs_pkey PRIMARY KEY (id);


--
-- Name: workouts workouts_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workouts
    ADD CONSTRAINT workouts_pkey PRIMARY KEY (id);


--
-- Name: idx_episode_tracking_episode_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episode_tracking_episode_id ON public.episode_tracking USING btree (episode_id);


--
-- Name: idx_episode_tracking_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episode_tracking_user_id ON public.episode_tracking USING btree (user_id);


--
-- Name: idx_episodes_external_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_external_id ON public.episodes USING btree (external_id);


--
-- Name: idx_episodes_filler; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_filler ON public.episodes USING btree (filler);


--
-- Name: idx_episodes_filler_recap; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_filler_recap ON public.episodes USING btree (filler, recap);


--
-- Name: idx_episodes_recap; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_recap ON public.episodes USING btree (recap);


--
-- Name: idx_episodes_show_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_show_id ON public.episodes USING btree (show_id);


--
-- Name: idx_market_data_last_updated; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_market_data_last_updated ON public.market_data USING btree (last_updated);


--
-- Name: idx_market_watchlist_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_market_watchlist_user_id ON public.market_watchlist USING btree (user_id);


--
-- Name: idx_medication_logs_date; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medication_logs_date ON public.medication_logs USING btree (date);


--
-- Name: idx_medication_logs_medication_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medication_logs_medication_id ON public.medication_logs USING btree (medication_id);


--
-- Name: idx_medication_logs_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medication_logs_user_id ON public.medication_logs USING btree (user_id);


--
-- Name: idx_medications_is_active; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medications_is_active ON public.medications USING btree (is_active);


--
-- Name: idx_medications_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medications_user_id ON public.medications USING btree (user_id);


--
-- Name: idx_news_articles_author; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_news_articles_author ON public.news_articles USING btree (author_name);


--
-- Name: idx_news_articles_category; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_news_articles_category ON public.news_articles USING btree (category);


--
-- Name: idx_news_articles_keywords; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_news_articles_keywords ON public.news_articles USING gin (to_tsvector('arabic'::regconfig, keywords));


--
-- Name: idx_news_articles_language; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_news_articles_language ON public.news_articles USING btree (language);


--
-- Name: idx_news_articles_published_at; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_news_articles_published_at ON public.news_articles USING btree (published_at DESC);


--
-- Name: idx_news_articles_source_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_news_articles_source_id ON public.news_articles USING btree (source_id);


--
-- Name: idx_notifications_created_at; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_notifications_created_at ON public.notifications USING btree (created_at DESC);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_notifications_user_id ON public.notifications USING btree (user_id);


--
-- Name: idx_notifications_user_read; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_notifications_user_read ON public.notifications USING btree (user_id, read);


--
-- Name: idx_projects_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_projects_user_id ON public.projects USING btree (user_id);


--
-- Name: idx_shows_external_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_external_id ON public.shows USING btree (external_id);


--
-- Name: idx_shows_external_id_type; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_external_id_type ON public.shows USING btree (external_id, show_type);


--
-- Name: idx_shows_last_episode_sync; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_last_episode_sync ON public.shows USING btree (last_episode_sync);


--
-- Name: idx_shows_last_info_sync; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_last_info_sync ON public.shows USING btree (last_info_sync);


--
-- Name: idx_shows_status; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_status ON public.shows USING btree (status);


--
-- Name: idx_shows_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_user_id ON public.shows USING btree (user_id);


--
-- Name: idx_sync_logs_created_at; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_sync_logs_created_at ON public.sync_logs USING btree (created_at);


--
-- Name: idx_sync_logs_show_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_sync_logs_show_id ON public.sync_logs USING btree (show_id);


--
-- Name: idx_task_attachments_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_attachments_task_id ON public.task_attachments USING btree (task_id);


--
-- Name: idx_task_comments_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_comments_task_id ON public.task_comments USING btree (task_id);


--
-- Name: idx_task_dependencies_depends_on_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_dependencies_depends_on_task_id ON public.task_dependencies USING btree (depends_on_task_id);


--
-- Name: idx_task_dependencies_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_dependencies_task_id ON public.task_dependencies USING btree (task_id);


--
-- Name: idx_tasks_completed; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_completed ON public.tasks USING btree (completed);


--
-- Name: idx_tasks_due_date; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_due_date ON public.tasks USING btree (due_date);


--
-- Name: idx_tasks_parent_display_order; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_parent_display_order ON public.tasks USING btree (parent_task_id, display_order) WHERE (parent_task_id IS NOT NULL);


--
-- Name: idx_tasks_parent_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_parent_task_id ON public.tasks USING btree (parent_task_id);


--
-- Name: idx_tasks_priority; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_priority ON public.tasks USING btree (priority);


--
-- Name: idx_tasks_project_display_order; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_project_display_order ON public.tasks USING btree (project_id, display_order);


--
-- Name: idx_tasks_project_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_project_id ON public.tasks USING btree (project_id);


--
-- Name: idx_tasks_status; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_status ON public.tasks USING btree (status);


--
-- Name: idx_tasks_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_user_id ON public.tasks USING btree (user_id);


--
-- Name: idx_user_interests_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_user_interests_user_id ON public.user_interests USING btree (user_id);


--
-- Name: idx_user_news_preferences_enabled; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_user_news_preferences_enabled ON public.user_news_preferences USING btree (user_id, is_enabled);


--
-- Name: idx_user_news_preferences_source_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_user_news_preferences_source_id ON public.user_news_preferences USING btree (source_id);


--
-- Name: idx_user_news_preferences_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_user_news_preferences_user_id ON public.user_news_preferences USING btree (user_id);


--
-- Name: user_news_preferences update_user_news_preferences_updated_at; Type: TRIGGER; Schema: public; Owner: most3mr
--

CREATE TRIGGER update_user_news_preferences_updated_at BEFORE UPDATE ON public.user_news_preferences FOR EACH ROW EXECUTE FUNCTION public.update_user_news_preferences_updated_at();


--
-- Name: episode_tracking episode_tracking_episode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episodes(id) ON DELETE CASCADE;


--
-- Name: episode_tracking episode_tracking_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: episodes episodes_show_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_show_id_fkey FOREIGN KEY (show_id) REFERENCES public.shows(id) ON DELETE CASCADE;


--
-- Name: episodes episodes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: habits_completions habits_completions_habit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits_completions
    ADD CONSTRAINT habits_completions_habit_id_fkey FOREIGN KEY (habit_id) REFERENCES public.habits(id) ON DELETE CASCADE;


--
-- Name: habits_completions habits_completions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits_completions
    ADD CONSTRAINT habits_completions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: habits habits_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits
    ADD CONSTRAINT habits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: market_watchlist market_watchlist_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.market_watchlist
    ADD CONSTRAINT market_watchlist_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: medication_logs medication_logs_medication_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_medication_id_fkey FOREIGN KEY (medication_id) REFERENCES public.medications(id) ON DELETE CASCADE;


--
-- Name: medication_logs medication_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: medications medications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medications
    ADD CONSTRAINT medications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mood_ratings mood_ratings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.mood_ratings
    ADD CONSTRAINT mood_ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: news_articles news_articles_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.news_articles
    ADD CONSTRAINT news_articles_source_id_fkey FOREIGN KEY (source_id) REFERENCES public.news_sources(id) ON DELETE CASCADE;


--
-- Name: notes notes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: projects projects_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: shows shows_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.shows
    ADD CONSTRAINT shows_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sync_logs sync_logs_show_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.sync_logs
    ADD CONSTRAINT sync_logs_show_id_fkey FOREIGN KEY (show_id) REFERENCES public.shows(id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_dependencies task_dependencies_depends_on_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_depends_on_task_id_fkey FOREIGN KEY (depends_on_task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_dependencies task_dependencies_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_parent_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_parent_task_id_fkey FOREIGN KEY (parent_task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: todos todos_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_interests user_interests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.user_interests
    ADD CONSTRAINT user_interests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_news_preferences user_news_preferences_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.user_news_preferences
    ADD CONSTRAINT user_news_preferences_source_id_fkey FOREIGN KEY (source_id) REFERENCES public.news_sources(id) ON DELETE CASCADE;


--
-- Name: user_news_preferences user_news_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.user_news_preferences
    ADD CONSTRAINT user_news_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: workout_logs workout_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workout_logs
    ADD CONSTRAINT workout_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: workouts workouts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workouts
    ADD CONSTRAINT workouts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

