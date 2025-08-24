-- Finance Management Feature Migration
-- Create comprehensive finance tracking tables

-- 1. User Financial Profiles
CREATE TABLE user_finances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    monthly_income DECIMAL(15,2) NOT NULL DEFAULT 0,
    emergency_buffer_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    emergency_buffer_percentage DECIMAL(5,2) NOT NULL DEFAULT 10.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

-- 2. Recurring Payments (loans, subscriptions, bills, medical)
CREATE TABLE recurring_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('loan', 'subscription', 'bill', 'medical', 'other')),
    amount DECIMAL(15,2) NOT NULL,
    frequency TEXT NOT NULL CHECK (frequency IN ('monthly', 'yearly', 'weekly', 'quarterly')),
    due_date INTEGER NOT NULL CHECK (due_date >= 1 AND due_date <= 31), -- Day of month for monthly/yearly, day of week for weekly
    start_date DATE NOT NULL,
    end_date DATE, -- NULL for ongoing subscriptions
    remaining_payments INTEGER, -- For loans, NULL for ongoing
    total_amount DECIMAL(15,2), -- Total loan amount for loans
    description TEXT,
    provider TEXT, -- Bank name, service provider, etc.
    account_number TEXT, -- Last 4 digits or identifier
    auto_pay BOOLEAN DEFAULT FALSE,
    renewal_notice_days INTEGER DEFAULT 30, -- Days before renewal to notify
    price_history JSONB DEFAULT '[]', -- Track price changes over time
    metadata JSONB DEFAULT '{}', -- Store additional data like loan interest rate, subscription features, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- 3. Savings Goals
CREATE TABLE savings_goals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    target_amount DECIMAL(15,2) NOT NULL,
    current_amount DECIMAL(15,2) DEFAULT 0,
    target_date DATE,
    priority TEXT NOT NULL CHECK (priority IN ('essential', 'important', 'nice-to-have')),
    monthly_contribution DECIMAL(15,2) DEFAULT 0,
    auto_calculate_contribution BOOLEAN DEFAULT TRUE,
    product_url TEXT, -- Link to product if applicable
    image_url TEXT, -- Product image
    description TEXT,
    is_achieved BOOLEAN DEFAULT FALSE,
    achieved_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- 4. Payment Logs (track when payments are made)
CREATE TABLE payment_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recurring_payment_id UUID REFERENCES recurring_payments(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    payment_date DATE NOT NULL,
    due_date DATE NOT NULL,
    payment_method TEXT, -- cash, card, bank_transfer, auto_pay
    notes TEXT,
    is_late BOOLEAN DEFAULT FALSE,
    late_fee DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- 5. Savings Contributions (track savings deposits)
CREATE TABLE savings_contributions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    savings_goal_id UUID NOT NULL REFERENCES savings_goals(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    contribution_date DATE NOT NULL,
    method TEXT, -- manual, automatic, bonus
    notes TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- 6. Financial Analytics Cache (for dashboard performance)
CREATE TABLE finance_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month_year VARCHAR(7) NOT NULL, -- Format: YYYY-MM
    total_income DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_fixed_payments DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_savings_contributions DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_emergency_buffer DECIMAL(15,2) NOT NULL DEFAULT 0,
    safe_to_spend DECIMAL(15,2) NOT NULL DEFAULT 0,
    payments_completed INTEGER DEFAULT 0,
    payments_remaining INTEGER DEFAULT 0,
    goals_achieved INTEGER DEFAULT 0,
    calculated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, month_year)
);

-- Indexes for better performance
CREATE INDEX idx_recurring_payments_user_id ON recurring_payments(user_id);
CREATE INDEX idx_recurring_payments_due_date ON recurring_payments(due_date);
CREATE INDEX idx_recurring_payments_category ON recurring_payments(category);
CREATE INDEX idx_recurring_payments_active ON recurring_payments(is_active);

CREATE INDEX idx_savings_goals_user_id ON savings_goals(user_id);
CREATE INDEX idx_savings_goals_priority ON savings_goals(priority);
CREATE INDEX idx_savings_goals_active ON savings_goals(is_active);
CREATE INDEX idx_savings_goals_target_date ON savings_goals(target_date);

CREATE INDEX idx_payment_logs_user_id ON payment_logs(user_id);
CREATE INDEX idx_payment_logs_payment_date ON payment_logs(payment_date);
CREATE INDEX idx_payment_logs_recurring_payment_id ON payment_logs(recurring_payment_id);

CREATE INDEX idx_savings_contributions_user_id ON savings_contributions(user_id);
CREATE INDEX idx_savings_contributions_goal_id ON savings_contributions(savings_goal_id);
CREATE INDEX idx_savings_contributions_date ON savings_contributions(contribution_date);

CREATE INDEX idx_finance_analytics_user_month ON finance_analytics(user_id, month_year);

-- Add triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_finance_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_finances_updated_at 
    BEFORE UPDATE ON user_finances 
    FOR EACH ROW EXECUTE FUNCTION update_finance_updated_at_column();

CREATE TRIGGER update_recurring_payments_updated_at 
    BEFORE UPDATE ON recurring_payments 
    FOR EACH ROW EXECUTE FUNCTION update_finance_updated_at_column();

CREATE TRIGGER update_savings_goals_updated_at 
    BEFORE UPDATE ON savings_goals 
    FOR EACH ROW EXECUTE FUNCTION update_finance_updated_at_column();