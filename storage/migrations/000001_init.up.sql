create table if not exists user_account (
    id                 serial       primary key,
    email              varchar(100)     null constraint unique_email unique,
    username           varchar(30)  not null,
    password           varchar(100)     null,
    first_name         varchar(50)      null,
    last_name          varchar(50)      null,
    created            timestamp    not null default current_timestamp,
    modified           timestamp    not null default current_timestamp,
    human_score        numeric(3,2)     null,
    confirmation       varchar(255)     null,
    sign_off           timestamp        null,
    sign_off_feedback  text             null,
    notification       boolean      not null default true,
    access_role        varchar(20)  not null default 'USER'
);

create index idx_user_confirmation on user_account (Confirmation);
create index idx_user_role on user_account (access_role);

create table if not exists email_message_sent (
    id        serial       primary key,
    recipient varchar(100) not null,
    subject   varchar(200) not null,
    body      text         not null,
    username  varchar(30)      null,
    sent      timestamp        null default current_timestamp
);

create table if not exists sign_in_attempt (
    id            serial       primary key,
    identifier    varchar(100) not null,
    human_score   numeric(3,2) not null,
    created       timestamp    not null default current_timestamp,
    status        varchar(10)  not null,
    ip_address    varchar(30)      null,
    failed_match  varchar(15)      null
);

create index idx_sign_in_status on sign_in_attempt (status);
create index idx_ip_address on sign_in_attempt (ip_address);