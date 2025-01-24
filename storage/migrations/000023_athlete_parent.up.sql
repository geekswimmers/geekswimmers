create table if not exists athlete (
    id           serial      primary key,
    first_name   text        not null,
    last_name    text        not null,
    birth_date   date        not null,
    gender       varchar(10) not null,
    user_account integer     null references user_account,
    created      timestamp   not null default current_timestamp
);

create table if not exists parent_athlete (
    id      serial  primary key,
    parent  integer not null references user_account,
    athlete integer not null references athlete
);

create table if not exists event_athlete (
    id           serial    primary key,
    event        integer   not null references swim_style_distance,
    athlete      integer   not null references athlete,
    best_time    integer   not null,
    last_update  timestamp not null default current_timestamp
);