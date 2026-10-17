alter table time_standard rename to standard;
alter table meet rename column time_standard to standard;
alter table standard_time rename column standard to standard_time;
alter table standard_time rename column time_standard to standard;