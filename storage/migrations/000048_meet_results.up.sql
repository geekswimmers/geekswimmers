-- meet_event: a specific event within a meet (e.g. Women 13-14 100m Freestyle)
create table meet_event (
    id      serial      primary key,
    meet    integer     not null references meet(id),
    event   integer     not null references swim_event(id),
    gender  varchar(10) not null  -- MALE, FEMALE
);

create unique index idx_meet_event on meet_event (meet, event, gender);

-- meet_result: a swimmer's official result in a meet_event.
-- Only swimmers and teams already registered in the system are imported.
-- result_time is null when the swimmer was disqualified. Scratched entries are not imported.
create table meet_result (
    id          serial      primary key,
    meet_event  integer     not null references meet_event(id),
    swimmer     integer     not null references swimmer(id),
    team        integer     not null references team(id),
    result_time integer         null,
    dq          boolean     not null default false
);

create unique index idx_meet_result on meet_result (meet_event, swimmer);
