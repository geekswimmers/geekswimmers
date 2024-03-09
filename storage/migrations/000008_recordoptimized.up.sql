alter table record_definition add min_age integer null;
alter table record_definition add max_age integer null;
alter table record_definition drop column age;

create index idx_def_fields on record_definition (gender, course, stroke, distance);
create index idx_def_age on record_definition (min_age, max_age);