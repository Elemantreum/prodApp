DROP TRIGGER training_capacity_changed ON trainings;
DROP FUNCTION training_waitlist_capacity_changed();
DROP TRIGGER booking_waitlist_changed ON bookings;
DROP FUNCTION update_booking_waitlist();
DROP FUNCTION notify_training_waitlist(integer);
DROP TABLE training_waitlist;
DROP TRIGGER lock_waitlist_on_booking_change ON bookings;
DROP FUNCTION lock_booking_training();
