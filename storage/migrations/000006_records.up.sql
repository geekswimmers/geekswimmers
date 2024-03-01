create table if not exists jurisdiction (
	id       serial      primary key,
	country  varchar(50) not null, -- CANADA
	province varchar(50)     null, -- ONTARIO
	region   varchar(50)     null, -- WESTERN
    city     varchar(50)     null, -- WATERLOO
    club     varchar(50)     null, -- ROW
    meet     varchar(50)     null  -- DEAN_BOLE, CUNNINGHAM_CLASSIC, NOVICE_1, NOVICE_2
);

create table if not exists record (
    id           serial       primary key,
    jurisdiction integer      not null references jurisdiction,
    age          integer          null,
    gender       varchar(10)      null, -- MALE, FEMALE
    course       varchar(10)      null, -- LONG, SHORT
    stroke       varchar(10)      null, -- FREE, BREAST, BACK, FLY, MEDLEY
    distance     integer          null
);

create table if not exists record_time (
    id          serial      primary key,
    record      integer     not null references record,
    holder      varchar(50)     null,
    record_time integer     not null,
    record_date date            null
);

alter table time_standard add jurisdiction integer null references jurisdiction;