create unique index uidx_swim_stroke on swim_style (stroke);
update standard_time set style = 'BREASTSTROKE' where style = 'BREAST';
alter table standard_time add constraint fk_standard_time_style foreign key (style) references swim_style (stroke);
alter table record_definition add constraint fk_record_definition_style foreign key (style) references swim_style (stroke);
