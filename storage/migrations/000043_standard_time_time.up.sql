alter table standard_time add if not exists update_date date null;

drop index public.idx_standard_time;
