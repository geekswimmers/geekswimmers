alter table athlete rename to swimmer;
drop table event_athlete;
alter table parent_athlete rename to parent_swimmer;
alter table parent_swimmer rename column athlete to swimmer;

update user_account set access_role = 'SWIMMER' where access_role = 'ATHLETE';