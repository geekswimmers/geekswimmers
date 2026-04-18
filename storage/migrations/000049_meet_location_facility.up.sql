alter table meet add column if not exists location varchar(100) null;
alter table meet add column if not exists facility varchar(100) null;
alter table meet alter column season drop not null;
