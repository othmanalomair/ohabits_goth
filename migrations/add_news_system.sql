-- Migration: Add News & Market Intelligence Dashboard
-- Date: 2025-01-14
-- Description: Creates tables for news aggregation, market data tracking, and user interests

-- News sources configuration
CREATE TABLE IF NOT EXISTS public.news_sources (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    url text NOT NULL,
    language text NOT NULL,
    category text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    fetch_frequency_hours integer DEFAULT 2 NOT NULL,
    last_fetched timestamp without time zone,
    created_at timestamp without time zone DEFAULT now()
);

-- News articles storage
CREATE TABLE IF NOT EXISTS public.news_articles (
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
    updated_at timestamp without time zone DEFAULT now()
);

-- User interests and customization
CREATE TABLE IF NOT EXISTS public.user_interests (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    interest_name text NOT NULL,
    keywords jsonb,
    sources jsonb,
    priority integer DEFAULT 1 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);

-- Market watchlist for users
CREATE TABLE IF NOT EXISTS public.market_watchlist (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    symbol text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);

-- Market data cache
CREATE TABLE IF NOT EXISTS public.market_data (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    symbol text NOT NULL,
    current_price numeric(15,6),
    change_amount numeric(15,6),
    change_percent numeric(8,4),
    volume bigint,
    market_cap bigint,
    last_updated timestamp without time zone DEFAULT now()
);

-- Primary key constraints
ALTER TABLE public.news_sources ADD CONSTRAINT news_sources_pkey PRIMARY KEY (id);
ALTER TABLE public.news_articles ADD CONSTRAINT news_articles_pkey PRIMARY KEY (id);
ALTER TABLE public.user_interests ADD CONSTRAINT user_interests_pkey PRIMARY KEY (id);
ALTER TABLE public.market_watchlist ADD CONSTRAINT market_watchlist_pkey PRIMARY KEY (id);
ALTER TABLE public.market_data ADD CONSTRAINT market_data_pkey PRIMARY KEY (id);

-- Unique constraints
ALTER TABLE public.news_articles ADD CONSTRAINT news_articles_url_unique UNIQUE (original_url);
ALTER TABLE public.market_data ADD CONSTRAINT market_data_symbol_unique UNIQUE (symbol);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_news_articles_source_id ON public.news_articles USING btree (source_id);
CREATE INDEX IF NOT EXISTS idx_news_articles_published_at ON public.news_articles USING btree (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_news_articles_category ON public.news_articles USING btree (category);
CREATE INDEX IF NOT EXISTS idx_news_articles_language ON public.news_articles USING btree (language);
CREATE INDEX IF NOT EXISTS idx_user_interests_user_id ON public.user_interests USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_market_watchlist_user_id ON public.market_watchlist USING btree (user_id);
CREATE INDEX IF NOT EXISTS idx_market_data_last_updated ON public.market_data USING btree (last_updated);

-- Foreign key constraints
ALTER TABLE public.news_articles ADD CONSTRAINT news_articles_source_id_fkey FOREIGN KEY (source_id) REFERENCES public.news_sources(id) ON DELETE CASCADE;
ALTER TABLE public.user_interests ADD CONSTRAINT user_interests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE public.market_watchlist ADD CONSTRAINT market_watchlist_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- Check constraints
ALTER TABLE public.news_sources ADD CONSTRAINT news_sources_type_check CHECK ((type = ANY (ARRAY['rss'::text, 'api'::text, 'scraper'::text])));
ALTER TABLE public.news_sources ADD CONSTRAINT news_sources_language_check CHECK ((language = ANY (ARRAY['ar'::text, 'en'::text])));
ALTER TABLE public.news_articles ADD CONSTRAINT news_articles_language_check CHECK ((language = ANY (ARRAY['ar'::text, 'en'::text])));
ALTER TABLE public.user_interests ADD CONSTRAINT user_interests_priority_check CHECK (((priority >= 1) AND (priority <= 5)));
ALTER TABLE public.market_watchlist ADD CONSTRAINT market_watchlist_type_check CHECK ((type = ANY (ARRAY['stock'::text, 'crypto'::text, 'forex'::text])));

-- Insert default Kuwait news sources
INSERT INTO public.news_sources (name, type, url, language, category, is_active, fetch_frequency_hours) VALUES
('Al-Jarida', 'rss', 'https://www.aljarida.com/rssFeed/1', 'ar', 'kuwait', true, 2),
('Al-Rai', 'rss', 'https://www.alraimedia.com/rssFeed/1', 'ar', 'kuwait', true, 2)
ON CONFLICT DO NOTHING;