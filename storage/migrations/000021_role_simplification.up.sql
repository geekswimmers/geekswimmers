drop table user_role;
alter table user_account add if not exists access_role varchar(20) not null default 'PARENT';
alter table sign_in_attempt alter failed_match type varchar(20);