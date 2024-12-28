-- defines a report
create table if not exists report (
    id          serial  primary key,
    name        text    not null
);

-- defines where values are placed on the report
create table if not exists report_mapping (
    id          serial      primary key,
    report      integer     not null references report,
    placeholder varchar(50) not null,
    coord_x     integer     not null,
    coord_y     integer     not null
);

-- defines the values to be placed on the report
create table if not exists record_poster (
    id      serial      primary key,
    mapping integer     not null references report_mapping,
    record  integer     not null references record,
    field   varchar(30) not null
);