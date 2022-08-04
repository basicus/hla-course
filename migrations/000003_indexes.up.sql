begin;

ALTER TABLE users DROP INDEX users_name_idx;
ALTER TABLE user_friend DROP INDEX user_friends_uid_idx;
ALTER TABLE user_friend DROP INDEX user_friends_fid_idx;
alter table users DROP INDEX users_login_idx;
alter table users DROP INDEX users_country_idx;
ALTER TABLE users ADD INDEX  users_name_idx (name, surname) USING BTREE;
ALTER TABLE user_friend ADD INDEX user_friends_uid_idx (user_id)  USING BTREE;
ALTER TABLE user_friend ADD INDEX user_friends_fid_idx (friend_id)  USING BTREE;
ALTER TABLE users ADD INDEX users_login_idx (login)  USING BTREE;
ALTER TABLE users ADD INDEX users_country_idx (country)  USING BTREE;

commit;
