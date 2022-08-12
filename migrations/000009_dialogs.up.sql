begin;

create table if not exists chats
(
    id         bigint primary key AUTO_INCREMENT auto_increment,
    title      varchar(200)                       not null collate utf8_general_ci,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP not null,
    closed     bool     default false             not null
);

ALTER TABLE chats
    ADD INDEX chats_users_idx (created_at) USING BTREE;


create table if not exists chat_participants
(
    id      bigint primary key AUTO_INCREMENT auto_increment,
    chat_id bigint,
    user_id bigint,
    status  int8 default 1 not null
);

alter table chat_participants
    add index chat_participants_idx (chat_id, user_id, status) using btree;

create table if not exists messages
(
    id        bigint primary key AUTO_INCREMENT auto_increment,
    chat_id   bigint        not null,
    user_from bigint        not null,
    send_at   datetime      not null,
    message   varchar(2000) not null collate utf8_general_ci

);

alter table messages
    add index chat_messages_idx (chat_id, send_at) using btree;

commit;