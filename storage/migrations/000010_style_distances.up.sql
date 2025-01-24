alter table swim_modality rename to swim_style;
alter table swim_modality_instruction rename to swim_style_instruction;
alter table swim_style_instruction rename column modality to style;
alter table record_definition rename column modality to style;
alter table standard_time rename column modality to style;

create table if not exists swim_style_distance (
    id          serial      primary key,
    style       integer     not null references swim_style,
    distance    integer     not null
);