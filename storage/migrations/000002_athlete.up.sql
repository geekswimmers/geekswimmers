create table if not exists athlete (
	id           serial      primary key,
	user_account integer     not null references user_account,
	date_birth   date        not null,
	gender       varchar(20) not null -- MALE, FEMALE
);

create table if not exists parent_athlete (
    id      serial primary key,
    parent  integer not null references user_account,
    athlete integer not null references athlete
);

create table if not exists club (
	id   serial       primary key,
	name varchar(100) not null
);

create table if not exists club_coach (
	id    serial  primary key,
	coach integer not null references user_account,
	club  integer not null references club
);

create table if not exists swim_season (
	id         serial      primary key,
	club       integer     not null references club,
	name       varchar(50) not null,
	start_date date        not null,
	end_date   date        not null
);

create table if not exists swim_group (
	id     serial       primary key,
	name   varchar(100) not null,
	club   integer      not null references club,
	season integer      not null references swim_season
);

create table if not exists swim_group_schedule (
    id         serial    primary key,
    swim_group integer   not null references swim_group,
    day_week   integer   not null,
    start_time timestamp not null,
    end_time   timestamp not null
);

create table if not exists swim_group_coach (
	id         serial  primary key,
	coach      integer not null references user_account,
	swim_group integer not null references swim_group
);

create table if not exists swim_group_athlete (
	id         serial  primary key,
	athlete    integer not null references athlete,
	swim_group integer not null references swim_group,
    start_date date        null,
	end_date   date        null
);

create table if not exists meet (
	id         serial       primary key,
	name       varchar(100) not null,
	course     varchar(10)  not null, -- LONG, SHORT, OPEN_WATER
    start_time timestamp        null,
    end_time   timestamp        null,
    organizer  integer          null references club
);

create table if not exists age_range (
    id          serial      primary key,
	description varchar(50) null,
	age_from    integer     null,
	age_to      integer     null
);

create table if not exists meet_session (
    if         serial        primary key,
    meet       integer       not null references meet,
    name       varchar(50)   not null,
    start_time timestamp         null,
    end_time   timestamp         null,
	session_type varchar(20)     null -- PRELIMINARY, SEMIFINAL, FINAL
);

create table if not exists meet_event (
	id       serial       primary key,
	session  integer      not null references meet_session,
	stroke   varchar(20)  not null, -- FREE, BREAST, BACK, BUTTERFLY, MEDLEY
	distance integer      not null,
	gender   varchar(20)      null -- FEMALE, MAKE, MIXED
);

create table if not exists meet_heat (
	id          serial  primary key,
	meet_event  integer not null references meet_event,
	heat_number integer not null
);

create table if not exists meet_athlete (
	id         serial primary key,
	athlete    integer not null references athlete,
	meet_heat  integer not null references meet_heat,
	lane       integer     null,
	race_time  integer     null
);

create table if not exists position (
    id          serial      primary key,
    name        varchar(50) not null,
    description text            null
);

create table if not exists meet_position (
    id              serial  primary key,
    volunteer       integer not null references user_account,
    meet_session    integer not null references meet_session,
    position        integer not null references position,
    deck_evaluation boolean not null default false,
    absent          boolean not null default false
);