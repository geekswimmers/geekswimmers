-- postgres trigger to convert time in mm:ss:ms to milliseconds
create or replace function convert_to_milliseconds(time_in_mmssms text) returns integer as $$
declare
    time_parts text[];
begin
    time_parts[1] := split_part(time_in_mmssms, ':', 1);
   	time_parts[2] := split_part(time_in_mmssms, ':', 2);
	time_parts[3] := split_part(time_in_mmssms, ':', 3);
    return (time_parts[1]::integer * 60000) + (time_parts[2]::integer * 1000) + (time_parts[3]::integer * 10);
end;
$$ language plpgsql;

-- trigger for standard_time
create or replace function standard_formatted_change() returns trigger as $$
begin
	update standard_time set standard = convert_to_milliseconds(new.standard_formatted) where id = new.id;
   	return null;
end;
$$ language plpgsql;

create trigger standard_formatted_changed 
after insert or update of standard_formatted on standard_time 
for each row execute function standard_formatted_change();

-- trigger for record
create or replace function record_formatted_change() returns trigger as $$
begin
	update record set record_time = convert_to_milliseconds(new.record_time_formatted) where id = new.id;
   return null;
end;
$$ language plpgsql;

create trigger record_time_formatted_changed 
after insert or update of record_time_formatted on record 
for each row execute function record_formatted_change();