alter table swimmer alter column birth_date drop not null;
alter table record add column swimmer integer references swimmer(id);