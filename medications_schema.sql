-- Medications tables for tracking medication schedules and logs

-- Medications table - stores medication information
CREATE TABLE public.medications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    name text NOT NULL,
    dosage text NOT NULL, -- e.g., "125mg", "1 tablet", "5ml"
    scheduled_days jsonb NOT NULL, -- same format as habits: ["monday", "tuesday", "daily"] or ["sun", "tue", "thu"]
    times_per_day integer NOT NULL DEFAULT 1, -- how many times per day (1, 2, 3, etc.)
    time_intervals text[], -- specific times: ["08:00", "20:00"] or ["morning", "evening"]
    duration_type text NOT NULL CHECK (duration_type IN ('lifetime', 'limited')),
    start_date date NOT NULL,
    end_date date, -- NULL for lifetime medications
    notes text, -- additional notes about the medication
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- Medication logs table - tracks when medications are taken
CREATE TABLE public.medication_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    medication_id uuid,
    user_id uuid,
    taken boolean NOT NULL,
    scheduled_time text, -- "08:00", "morning", "evening", etc.
    actual_time timestamp without time zone, -- when actually taken
    date date NOT NULL,
    notes text, -- optional notes for this specific log entry
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- Add primary keys
ALTER TABLE ONLY public.medications
    ADD CONSTRAINT medications_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_pkey PRIMARY KEY (id);

-- Add foreign key constraints
ALTER TABLE ONLY public.medications
    ADD CONSTRAINT medications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_medication_id_fkey FOREIGN KEY (medication_id) REFERENCES public.medications(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- Add indexes for better performance
CREATE INDEX idx_medications_user_id ON public.medications(user_id);
CREATE INDEX idx_medications_is_active ON public.medications(is_active);
CREATE INDEX idx_medication_logs_user_id ON public.medication_logs(user_id);
CREATE INDEX idx_medication_logs_date ON public.medication_logs(date);
CREATE INDEX idx_medication_logs_medication_id ON public.medication_logs(medication_id);

-- Set ownership (adjust as needed)
ALTER TABLE public.medications OWNER TO most3mr;
ALTER TABLE public.medication_logs OWNER TO most3mr;