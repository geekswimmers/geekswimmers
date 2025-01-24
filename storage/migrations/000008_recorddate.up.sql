alter table record add year integer null;
alter table record add month integer null;

update record set year = date_part('year', record_date) where record_date is not null;
update record set month = date_part('month', record_date) where record_date is not null;

alter table record drop column record_date;

update article set highlighted = false;