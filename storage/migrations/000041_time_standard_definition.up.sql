create table if not exists standard_definition (
    id            serial      primary key,
    age           integer     not null,
    gender        varchar(20) not null, -- MALE, FEMALE
    course        varchar(10) not null, -- LONG, SHORT
    style         varchar(20) not null, -- FREE, BREAST, BACK, FLY, MEDLEY
	distance      integer     not null
);

create unique index idx_standard_definition on standard_definition (age, gender, course, style, distance);