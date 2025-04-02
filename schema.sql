--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2
-- Dumped by pg_dump version 17.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
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
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.workouts OWNER TO most3mr;

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

