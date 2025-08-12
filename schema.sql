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
    tvmaze_id integer NOT NULL,
    name text NOT NULL,
    season integer NOT NULL,
    number integer NOT NULL,
    summary text,
    airdate date,
    runtime integer,
    image_url text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
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
-- Name: shows; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.shows (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    tvmaze_id integer NOT NULL,
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
    last_info_sync timestamp without time zone
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
-- Name: todos todos_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_pkey PRIMARY KEY (id);


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
-- Name: idx_episodes_show_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_show_id ON public.episodes USING btree (show_id);


--
-- Name: idx_episodes_tvmaze_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_tvmaze_id ON public.episodes USING btree (tvmaze_id);


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
-- Name: idx_shows_tvmaze_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_tvmaze_id ON public.shows USING btree (tvmaze_id);


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
-- Name: todos todos_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


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

