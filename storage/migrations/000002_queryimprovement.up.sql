drop index if exists idx_meet_age;

alter table meet add start_date date null;
alter table meet add end_date date null;
alter table meet add website varchar(100) null;
alter table meet add min_age_enforced boolean null default false;
alter table meet add max_age_enforced boolean null default false;

create index idx_upcoming_meets on meet (course, start_date);

alter table time_standard drop column summary;
alter table time_standard add min_age_time integer null;
alter table time_standard add max_age_time integer null;