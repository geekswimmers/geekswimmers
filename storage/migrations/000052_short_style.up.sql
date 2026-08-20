alter table swim_style add column stroke_short varchar(20) null;

update swim_style set stroke_short = 'FREE' where stroke = 'FREESTYLE';
update swim_style set stroke_short = 'BACK' where stroke = 'BACKSTROKE';
update swim_style set stroke_short = 'BREAST' where stroke = 'BREASTSTROKE';
update swim_style set stroke_short = 'FLY' where stroke = 'BUTTERFLY';
update swim_style set stroke_short = 'IM' where stroke = 'MEDLEY';
update swim_style set stroke_short = 'FREE_RELAY' where stroke = 'FREESTYLE_RELAY';
update swim_style set stroke_short = 'MEDLEY_RELAY' where stroke = 'MEDLEY_RELAY';