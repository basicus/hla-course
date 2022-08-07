begin;

create table if not exists posts
(
    id        bigint primary key AUTO_INCREMENT auto_increment,
    user_id   bigint,
    title     varchar(200) not null,
    message   varchar(2048),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME ON UPDATE CURRENT_TIMESTAMP,
    deleted bool
);

ALTER TABLE posts ADD INDEX  posts_users_idx (user_id, created_at) USING BTREE;

commit;
