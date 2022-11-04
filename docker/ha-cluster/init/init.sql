CREATE USER IF NOT EXISTS 'replication'@'%'  IDENTIFIED BY 'pass';
CREATE USER IF NOT EXISTS 'haproxy'@'haproxy.ha-cluster_default' IDENTIFIED BY 'password';
GRANT REPLICATION SLAVE ON *.* TO 'replication'@'%';
