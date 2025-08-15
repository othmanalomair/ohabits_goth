# Personal News & Market Intelligence Dashboard - Implementation Plan

## Project Overview
Build a personalized news aggregator and market tracker that fetches Kuwait local news (Arabic), global news (English), and financial market data, with AI-powered summaries and insights. This is for personal use (1-5 users).

## Architecture Integration
- **Framework**: Go with Gorilla Mux (following existing ohabits structure)
- **Database**: PostgreSQL (extend existing schema)  
- **Frontend**: HTMX + Server-side templates (consistent with current app)
- **Background Jobs**: Go routines for periodic data fetching
- **Authentication**: Use existing JWT-based auth system

## Phase 1: Kuwait Local News (CURRENT FOCUS)

### Database Schema Extensions
```sql
-- News sources configuration
CREATE TABLE news_sources (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'rss', 'api', 'scraper'
    url TEXT NOT NULL,
    language TEXT NOT NULL, -- 'ar', 'en'
    category TEXT NOT NULL, -- 'kuwait', 'global', 'tech', 'anime', etc.
    is_active BOOLEAN DEFAULT true,
    fetch_frequency_hours INTEGER DEFAULT 2,
    last_fetched TIMESTAMP,
    created_at TIMESTAMP DEFAULT now()
);

-- News articles storage
CREATE TABLE news_articles (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    source_id UUID REFERENCES news_sources(id),
    title TEXT NOT NULL,
    content TEXT,
    summary TEXT, -- AI-generated summary
    original_url TEXT NOT NULL,
    image_url TEXT,
    published_at TIMESTAMP,
    language TEXT NOT NULL,
    category TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

-- User interests and customization
CREATE TABLE user_interests (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    interest_name TEXT NOT NULL,
    keywords JSONB, -- ["anime", "technology", "kuwait"]
    sources JSONB, -- source IDs or RSS URLs
    priority INTEGER DEFAULT 1, -- 1-5 scale
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now()
);

-- Market watchlist for users
CREATE TABLE market_watchlist (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL, -- "AAPL", "BTC-USD", etc.
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'stock', 'crypto', 'forex'
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT now()
);

-- Market data cache
CREATE TABLE market_data (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    symbol TEXT NOT NULL,
    current_price DECIMAL(15,6),
    change_amount DECIMAL(15,6),
    change_percent DECIMAL(8,4),
    volume BIGINT,
    market_cap BIGINT,
    last_updated TIMESTAMP DEFAULT now()
);
```

### Kuwait News Sources (Phase 1 Implementation)
1. **Al-Jarida** - Local News: `https://www.aljarida.com/rssFeed/1`
2. **Al-Rai** - Local News: `https://www.alraimedia.com/rssFeed/1`

### Implementation Steps (Phase 1)

#### Step 1: Database Setup ✅
- ✅ Add news-related tables to schema.sql
- ✅ Create migration script (`migrations/add_news_system.sql`)

#### Step 2: Kuwait News RSS Parser ✅
- ✅ Create `internal/services/news/` package
- ✅ Implement RSS parser for Arabic sources
- ✅ Handle Arabic text encoding properly
- ✅ Store articles in database with deduplication

#### Step 3: News Display Interface ✅  
- ✅ Create `/news` route and handler
- ✅ Design news template with Arabic RTL support
- ✅ Add news link to overlay menu
- ✅ Implement pagination and filtering

#### Step 4: Background News Fetching ✅
- ✅ Create background service to fetch news every 2 hours
- ✅ Implement graceful error handling and retry logic
- ✅ Add logging for monitoring fetch status

## ✅ PHASE 1 COMPLETE!

### What's Implemented:
1. **Complete Database Schema** - All tables for news, sources, and user interests
2. **Kuwait News RSS Parser** - Supports Al-Jarida and Al-Rai working RSS feeds
3. **News Dashboard Interface** - Full-featured news page with Arabic RTL support
4. **Background News Fetching** - Automatic fetching every 30 minutes
5. **News Sources Management** - View and manage news sources
6. **HTMX Integration** - Dynamic updates without page reloads

### Ready to Use:
- News Dashboard accessible at `/news`
- Pre-configured Kuwait news sources
- Automatic background news fetching
- Responsive design with mobile support
- Arabic text support with proper RTL layout

## Phase 2: Global News Sources (Future)
- NewsAPI.org integration (100 req/day limit)
- Reddit API for tech/anime subreddits
- Hacker News API 
- RSS feeds: BBC, Reuters, TechCrunch, The Verge, AnimeNewsNetwork

## Phase 3: Market Data Integration (Future)
- Real-time price display in top bar
- yfinance for stock data
- CoinGecko API for crypto prices
- Alpha Vantage integration (25 req/day limit)
- Update frequency: Every 15 minutes during market hours

## Phase 4: AI-Powered Features (Future)
- Article summarization using OpenAI API
- Sentiment analysis for market-related news
- Personalized news recommendations
- Trend detection and insights

## Technical Considerations

### Arabic Text Support
- UTF-8 encoding throughout the system
- RTL (Right-to-Left) CSS for Arabic content
- Arabic-aware text processing for summaries

### Caching Strategy
- Redis for market data caching (optional)
- Database-level caching for news articles
- TTL-based cache invalidation

### Rate Limiting & API Management
- Respect API rate limits with exponential backoff
- Queue-based processing for high-volume sources
- Fallback mechanisms for API failures

### Performance Optimization
- Indexing on frequently queried columns
- Lazy loading for news images
- Pagination for large result sets

## Environment Variables Needed
```env
# News APIs
NEWSAPI_KEY=your_newsapi_key
REDDIT_CLIENT_ID=your_reddit_client_id
REDDIT_CLIENT_SECRET=your_reddit_secret

# Market Data APIs  
ALPHA_VANTAGE_KEY=your_alpha_vantage_key
COINGECKO_API_KEY=your_coingecko_key

# AI Services (future)
OPENAI_API_KEY=your_openai_key
```

## File Structure
```
internal/
├── services/
│   ├── news/
│   │   ├── rss_parser.go
│   │   ├── fetcher.go
│   │   └── sources.go
│   ├── market/
│   │   ├── stocks.go
│   │   └── crypto.go
│   └── ai/
│       └── summarizer.go
├── handlers/
│   ├── news.go
│   └── market.go
└── db/
    ├── news.go
    └── market.go

templates/
├── news.html
└── partials/
    ├── news_item.html
    └── market_ticker.html
```

## Success Metrics
- Successfully fetch and display Kuwait news in Arabic
- News updates every 2 hours automatically  
- Clean, responsive interface with RTL support
- Zero data loss during RSS parsing
- Fast page load times (<2s for news feed)

---

*This plan will be executed in phases, starting with Kuwait news aggregation as the foundation.*