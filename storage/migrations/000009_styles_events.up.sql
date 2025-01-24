create table if not exists swim_modality (
    id          serial      primary key,
    stroke      varchar(20) not null,
    description text            null,
    sequence    integer         null
);

create unique index idx_modality_stroke on swim_modality (stroke);

create table if not exists swim_modality_instruction (
    id          serial      primary key,
    modality    integer     not null references swim_modality,
    instruction text        not null,
    sequence    integer     not null
);

alter table record_definition rename column stroke to modality;
alter table record_definition alter column modality type varchar(20);

update record_definition set modality = 'FREESTYLE' where modality = 'FREE';
update record_definition set modality = 'BACKSTROKE' where modality = 'BACK';
update record_definition set modality = 'BREASTSTROKE' where modality = 'BREAST';
update record_definition set modality = 'BUTTERFLY' where modality = 'FLY';

alter table standard_time rename column stroke to modality;

update standard_time set modality = 'FREESTYLE' where modality = 'FREE';
update standard_time set modality = 'BACKSTROKE' where modality = 'BACK';
update standard_time set modality = 'BREASTSTROKE' where modality = 'BREAST';
update standard_time set modality = 'BUTTERFLY' where modality = 'FLY';