begin ;

alter table users add column shard_id varchar(5) collate latin1_general_ci;
update users set shard_id = '00000' where shard_id is null;

commit