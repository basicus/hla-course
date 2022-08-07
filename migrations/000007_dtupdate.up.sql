begin;


alter table posts  alter column updated_at set default current_timestamp();
update posts set updated_at=created_at where updated_at is null;

commit;