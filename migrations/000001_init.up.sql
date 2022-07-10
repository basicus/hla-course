begin;

create table if not exists users
(
    user_id   bigint primary key AUTO_INCREMENT auto_increment,
    login     varchar(200)    not null,
    email     varchar(100),
    phone     varchar(15),
    password  varchar(64) not null,
    name      varchar(128),
    surname   varchar(128),
    age       int,
    sex       enum ('male','female'),
    country   varchar(128),
    city      varchar(128),
    interests varchar(2048)

);

create table if not exists user_friend
(
    user_id   bigint,
    friend_id bigint
);

commit;
