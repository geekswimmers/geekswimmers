alter table parent_swimmer add approval varchar(10) not null default 'PENDING';
update parent_swimmer set approval = 'ACCEPTED' where approved;
alter table parent_swimmer drop column approved;