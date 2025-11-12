create or replace function standard_formatted_change() returns trigger as $$
begin
    update standard_time set standard = coalesce(standard, convert_to_milliseconds(new.standard_formatted)) where id = new.id;
    return null;
end;
$$ language plpgsql;

create or replace function record_formatted_change() returns trigger as $$
begin
    update record set record_time = coalesce(record_time, convert_to_milliseconds(new.record_time_formatted)) where id = new.id;
    return null;
end;
$$ language plpgsql;