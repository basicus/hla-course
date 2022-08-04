begin;

ALTER TABLE users
    ADD INDEX users_name_idx (name, surname);
ALTER TABLE user_friend
    ADD index user_friends_uid_idx (user_id);
ALTER TABLE user_friend
    ADD index user_friends_fid_idx (friend_id);
alter table users
    ADD index users_login_idx (login);
alter table users
    add index users_country_idx (country);

commit;
