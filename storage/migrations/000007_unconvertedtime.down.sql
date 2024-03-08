drop trigger if exists standard_formatted_changed on standard_time;
drop trigger if exists record_time_formatted_changed ON record;

drop function if exists standard_formatted_change();
drop function if exists record_formatted_change();
drop function if exists convert_to_milliseconds(text);

alter table standard_time drop column  if exists standard_formatted;
alter table record drop column  if exists record_time_formatted;

alter table standard_time add column standard_formatted varchar(12) null;
alter table standard_time alter column standard drop not null;
alter table record add column record_time_formatted varchar(12) null;
alter table record alter column record_time drop not null;