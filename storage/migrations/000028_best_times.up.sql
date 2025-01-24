alter table swim_style_distance rename to swim_event;

-- Use cases:
--  - the best time is defined for the first time, as a baseline: only the
--    required columns are informed.
--  - the user wants to update a baseline best time: they can update the
--    baseline record as long as there is no other record for the same
--    event and course.
--  - the swimmer achieve a new best time in a meet: a new record is inserted
--    with the columns 'updated' and 'meet' defined. The baseline best time
--    can not be updated anymore.
--  - the user made a mistake while inserting the best time obtained in a meet:
--    the record can be updated, but only the column 'best_time' can be changed.
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