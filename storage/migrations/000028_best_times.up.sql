alter table swim_style_distance rename to swim_event;

create table if not exists swimmer_best_time (
    id        serial      primary key,
    swimmer   integer     not null references swimmer,
    event     integer     not null references swim_event,
    course    varchar(10) not null,
    best_time integer     not null,
    updated   date            null,
    meet      integer         null references meet
);

create unique index if not exists idx_best_time
    on swimmer_best_time (swimmer, event, course, meet) nulls not distinct;

-- If the swimmer's best time is older than this date then a new best time
-- needs to be attempted to meet the standard.
alter table meet add column if not exists standard_time_oldest date null;