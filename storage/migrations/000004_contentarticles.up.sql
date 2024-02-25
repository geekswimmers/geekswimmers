alter table article add sub_title varchar(255) null;
update article set content = concat('articles/' , reference, '.md');
alter table article alter content type varchar(100);