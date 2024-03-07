create table if not exists jurisdiction (
	id       serial      primary key,
	country  varchar(50) not null, -- CANADA
	province varchar(50)     null, -- ONTARIO
	region   varchar(50)     null, -- WESTERN
    city     varchar(50)     null, -- WATERLOO
    club     varchar(50)     null, -- ROW
    meet     varchar(50)     null  -- DEAN_BOLE, CUNNINGHAM_CLASSIC, NOVICE_1, NOVICE_2
);

create table if not exists record_definition (
    id           serial       primary key,
    age          integer          null, -- If null, then it's a open record.
    gender       varchar(10)  not null, -- MALE, FEMALE
    course       varchar(10)  not null, -- LONG, SHORT
    stroke       varchar(10)  not null, -- FREE, BREAST, BACK, FLY, MEDLEY
    distance     integer      not null
);

create table if not exists record (
    id           serial      primary key,
    jurisdiction integer         null references jurisdiction, -- If null, then it's a world record.
    definition   integer     not null references record_definition,
    record_time  integer     not null,
    record_date  date            null,
    holder       varchar(50)     null
);

alter table time_standard add jurisdiction integer null references jurisdiction;