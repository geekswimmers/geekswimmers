create table if not exists service_update (
    id        serial       primary key,
    title     varchar(255) not null,
    content   text             null,
    published date             null
);
