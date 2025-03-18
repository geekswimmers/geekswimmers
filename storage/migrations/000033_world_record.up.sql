alter table record add column cross_link int null; -- Link between records, for example a club record can also be a World Record.
alter table jurisdiction add column world varchar(50) null;
alter table jurisdiction alter column country drop not null;