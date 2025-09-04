-- Four swimmers per record to allow relay records.
alter table record add if not exists column swimmer_st int null references swimmer (id); -- First
alter table record add if not exists column swimmer_nd int null references swimmer (id); -- Second
alter table record add if not exists column swimmer_rd int null references swimmer (id); -- Third
alter table record add if not exists column swimmer_th int null references swimmer (id); -- Fourth