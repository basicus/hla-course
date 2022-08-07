begin;

ALTER DATABASE project CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
alter table posts CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
alter table users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
alter table posts modify title varchar(200) collate utf8mb4_unicode_ci not null;
alter table posts modify message varchar(2048) collate utf8mb4_unicode_ci null;
commit;


