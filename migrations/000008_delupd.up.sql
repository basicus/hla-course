begin;


alter table posts
    modify deleted bool default false;
update posts
set deleted= false
where deleted is null;
alter table posts
    modify deleted bool default false not null;

commit;