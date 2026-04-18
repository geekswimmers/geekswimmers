alter table swimmer add column if not exists national_number varchar(20) null;

create unique index if not exists idx_swimmer_national_number on swimmer (national_number);
