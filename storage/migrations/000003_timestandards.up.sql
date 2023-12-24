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