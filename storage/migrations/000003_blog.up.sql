create table if not exists article (
	reference    varchar(100) primary key,
	title        varchar(100) not null,
	abstract     text         not null,
	highlighted  boolean      not null default false,
	published    date         not null default current_date,
	content      text         not null
);

create index idx_article_published on article (highlighted, published);