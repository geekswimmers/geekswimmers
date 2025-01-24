create table if not exists quote (
    seq      integer      primary key,
    quote    varchar(255) not null,
    author   varchar(50)  null
);