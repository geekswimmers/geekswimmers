create table if not exists user_account (
    id                serial       primary key,
    first_name        varchar(50)  not null,
    last_name         varchar(50)  not null,
    gender            varchar(2)       null,
    birth_date        date             null,
    confirmation      varchar(255)     null,
    human_score       numeric(3,2)     null,
    email             varchar(100)     null,
    username          varchar(30)      null,
    password          varchar(100)     null,
    notification      boolean      not null default true,
    sign_off          timestamp        null,
    sign_off_feedback text             null,
    created           timestamp    not null current_timestamp,
    modified          timestamp    not null current_timestamp
);

create index idx_user_confirmation on user_account (confirmation);

create table if not exists user_role (
    id           serial      primary key,
    user_account integer     not null references user_account(id),
    role         varchar(20) not null default 'PARENT', -- ATHLETE, COACH, OFFICIAL
);

create index udx_user_role on user_role (user_account, role);

create table if not exists family (
    id     serial   primary key,
    member integer     not null references user_account(id),
    main   boolean     not null default false
);

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