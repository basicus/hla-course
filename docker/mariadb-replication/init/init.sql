CREATE USER IF NOT EXISTS 'replication'@'%'  IDENTIFIED BY 'pass';
GRANT REPLICATION SLAVE ON *.* TO 'replication'@'%';
