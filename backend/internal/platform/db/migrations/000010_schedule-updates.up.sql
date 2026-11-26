CREATE FUNCTION notify_schedule_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_notify('schedule_changed', '');
    RETURN NULL;
END;
$$;
CREATE TRIGGER notify_schedule_bookings AFTER INSERT OR UPDATE OR DELETE ON bookings
FOR EACH ROW EXECUTE FUNCTION notify_schedule_change();
CREATE TRIGGER notify_schedule_trainings AFTER INSERT OR UPDATE OR DELETE ON trainings
FOR EACH ROW EXECUTE FUNCTION notify_schedule_change();
