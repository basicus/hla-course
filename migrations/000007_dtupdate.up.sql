begin;


-- alter table posts alter column updated_at updated_at SET TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
alter table posts modify updated_at timestamp null on update CURRENT_TIMESTAMP;
update posts set updated_at=created_at where updated_at is null;

commit;