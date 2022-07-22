begin;

ALTER TABLE users
    ADD INDEX IF NOT EXISTS users_name_idx (name, surname);
ALTER TABLE user_friend
    ADD index if not exists user_friends_uid_idx (user_id);
ALTER TABLE user_friend
    ADD index if not exists user_friends_fid_idx (friend_id);
alter table users
    ADD index if not exists users_login_idx (login);
alter table users
    add index if not exists users_country_idx (country);

commit;
