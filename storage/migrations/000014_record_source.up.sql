create table if not exists record_set (
    id           serial       primary key,
    jurisdiction integer      null references jurisdiction,-- If null, then it's a world record.
    source_title varchar(100) null,
    source_link  varchar(255) null
);

insert into record_set (id, jurisdiction, source_title, source_link) values
    (1, 1, 'ROW Website', 'https://www.rowswimming.ca/page/for-swimmers/row-records'),
    (2, 2, 'Swim Ontario', 'https://www.swimontario.com/athletes/records');

alter table record add record_set integer null references record_set;

update record set record_set = (select id from record_set where jurisdiction = 1) where jurisdiction = 1;
update record set record_set = (select id from record_set where jurisdiction = 2) where jurisdiction = 2;

alter table record drop column jurisdiction;