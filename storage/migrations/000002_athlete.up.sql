create table if not exists athlete (
	id           serial      primary key,
	user_account integer     not null references user_account,
	birthday     date        not null,
	gender       varchar(20) not null
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

create table if not exists swim_group (
	id   serial       primary key,
	name varchar(100) not null,
	club integer      not null references club
);

create table if not exists swim_group_schedule (
    id         serial primary key,
    swim_group integer not null references swim_group,
    day_week   integer not null,
    start_time timestamp not null,
    end_time   timestamp not null
);

create table if not exists swim_group_coach (
	id         serial  primary key,
	coach      integer not null references user_account,
	swim_group integer not null references swim_group,
	start_date date    null,
	end_date   date    null
);

create table if not exists swim_group_athlete (
	id         serial  primary key,
	athlete    integer not null references athlete,
	swim_group integer not null references swim_group,
    start_date date    null,
	end_date   date    null
);

create table if not exists meet (
	id         serial       primary key,
	name       varchar(100) not null,
    start_time timestamp    null,
    end_time   timestamp    null,
    organizer  integer      null references club
);

create table if not exists meet_session (
    if         serial       primary key,
    meet       integer not  null references meet,
    name       varchar(50)  not null,
    start_time timestamp    null,
    end_time   timestamp    null
);

create table if not exists meet_event (
	id       serial       primary key,
	session  integer      not null references meet_session,
	stroke   varchar(20)  not null,
	distance integer      not null
);

create table if not exists meet_athlete (
	id         serial primary key,
	meet_event integer not null,
	athlete    integer not null references athlete
);

create table if not exists timekeeping (
	id         serial  primary key,
	athlete    integer not null references athlete,
	meet_event integer not null references meet_event,
	time_taken integer not null,
	date_taken date    null,
	meet       integer null references meet
);

create table if not exists meet_position (
    id          serial       primary key,
    name        varchar(50)  not null,
    description varchar(255) null
);

create table if not exists volunteer_meet (
    id              serial  primary key,
    volunteer       integer not null references user_account,
    meet_session    integer not null references meet_session,
    position        integer not null references meet_position,
    deck_evaluation boolean not null default false,
    absent          boolean not null default false
);