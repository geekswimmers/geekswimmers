alter table athlete rename to swimmer;
alter table athlete_best_time rename to swimmer_best_time;
alter table swimmer_best_time rename column athlete to swimmer;
alter table athlete_goal_time rename to swimmer_goal_time;
alter table swimmer_goal_time rename column athlete to swimmer;
alter table event_athlete rename to event_swimmer;
alter table event_swimmer rename column athlete to swimmer;
alter table parent_athlete rename to parent_swimmer;
alter table parent_swimmer rename column athlete to swimmer;

update user_account set access_role = 'SWIMMER' where access_role = 'ATHLETE';