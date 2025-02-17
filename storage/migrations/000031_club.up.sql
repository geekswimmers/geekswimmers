create table if not exists club (
    id                serial       primary key,
    jurisdiction      integer      not null references jurisdiction(id),
    full_name         varchar(100) not null,
    achronym          varchar(10)  not null,
    website           varchar(100)     null
);

alter table meet add column organizer integer null references club(id);

alter table swimmer add column club integer null references club(id);