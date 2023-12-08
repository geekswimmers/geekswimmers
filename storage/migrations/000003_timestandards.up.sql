create table if not exists time_standard (
    id     serial       primary key,
    season integer      not null references swim_season,
    name   varchar(100) not null
);

create table if not exists standard_time (
    id         serial      primary key,
    standard   integer     not null references time_standard,
    age        integer         null,
    gender     varchar(20) not null, -- MALE, FEMALE
    course     varchar(10) not null, -- LONG, SHORT, OPEN_WATER
    stroke     varchar(20) not null, -- FREE, BREAST, BACK, BUTTERFLY, MEDLEY
	distance   integer     not null,
    quali_time integer     not null
);