CREATE OR REPLACE FUNCTION set_session_created_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.created_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER session_created_at_trigger
BEFORE INSERT ON sessions
FOR EACH ROW
EXECUTE FUNCTION set_session_created_at();