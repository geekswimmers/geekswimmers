create table if not exists article (
	reference varchar(100) primary key,
	title     varchar(100) not null,
	published date         not null,
	content   text         not null
);

create index idx_article_published on article (published);