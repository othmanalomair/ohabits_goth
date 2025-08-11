--
-- Shows and Episodes Migration
-- Run this migration to add TV/Anime tracking functionality
--

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
    updated_at timestamp without time zone DEFAULT now()
);

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

--
-- Primary Keys
--

ALTER TABLE ONLY public.shows
    ADD CONSTRAINT shows_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_pkey PRIMARY KEY (id);

-- Add unique constraint for episode_id, user_id combination
ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_episode_user_unique UNIQUE (episode_id, user_id);

--
-- Foreign Key Constraints
--

ALTER TABLE ONLY public.shows
    ADD CONSTRAINT shows_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_show_id_fkey FOREIGN KEY (show_id) REFERENCES public.shows(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episodes(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Indexes for better performance
--

CREATE INDEX idx_shows_user_id ON public.shows(user_id);
CREATE INDEX idx_shows_tvmaze_id ON public.shows(tvmaze_id);
CREATE INDEX idx_episodes_show_id ON public.episodes(show_id);
CREATE INDEX idx_episodes_tvmaze_id ON public.episodes(tvmaze_id);
CREATE INDEX idx_episode_tracking_episode_id ON public.episode_tracking(episode_id);
CREATE INDEX idx_episode_tracking_user_id ON public.episode_tracking(user_id);