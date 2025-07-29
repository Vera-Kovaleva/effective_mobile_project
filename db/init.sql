CREATE TABLE IF NOT EXISTS subscriptions (
    service_name TEXT NOT NULL,
    month_cost INTEGER NOT NULL,
    user_id UUID NOT NULL,
    subs_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    subs_end_date DATE
);

CREATE INDEX idx_subscriptions ON subscriptions(user_id, service_name, subs_start_date, subs_end_date) include (month_cost);
