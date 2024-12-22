drop table user_role;
alter table user_account add if not exists access_role varchar(20) not null default 'PARENT';