CREATE TABLE IF NOT EXISTS subscriptions (
    service_name TEXT NOT NULL,
    month_cost INTEGER NOT NULL,
    id UUID PRIMARY KEY NOT NULL,
    subs_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    subs_end_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE OR REPLACE FUNCTION update_is_active()
RETURNS TRIGGER 
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.is_active := (NEW.subs_end_date IS NULL OR NEW.subs_end_date > CURRENT_DATE);
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_update_is_active BEFORE INSERT OR UPDATE ON subscriptions FOR EACH ROW EXECUTE FUNCTION update_is_active();

CREATE UNIQUE INDEX subscription_in_progress_unique ON subscriptions (id, service_name) WHERE is_active;