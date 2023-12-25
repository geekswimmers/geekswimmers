create table if not exists swim_season (
	id         serial      primary key,
	name       varchar(50) not null,
	start_date date        not null,
	end_date   date        not null
);

create index idx_season_period on swim_season (start_date, end_date);

create table if not exists time_standard (
    id      serial       primary key,
    season  integer      not null references swim_season,
    name    varchar(100) not null,
    summary varchar(255)     null
);

create table if not exists standard_time (
    id            serial      primary key,
    time_standard integer     not null references time_standard,
    age           integer         null,
    gender        varchar(20) not null, -- MALE, FEMALE
    course        varchar(10) not null, -- LONG, SHORT
    stroke        varchar(20) not null, -- FREE, BREAST, BACK, FLY, MEDLEY
	distance      integer     not null,
    standard      integer     not null
);

create unique index idx_standard_time on standard_time (time_standard, age, gender, course, stroke, distance);

create table if not exists meet (
	id            serial       primary key,
	name          varchar(100) not null,
	course        varchar(10)  not null, -- LONG, SHORT
    season        integer      not null references swim_season,
    time_standard integer          null references time_standard,
    age_date      timestamp        null
);

create index idx_meet_age on meet (course);