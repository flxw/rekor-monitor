CREATE DATABASE rekor;
CREATE USER 'grafana'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON rekor.* TO 'grafana'@'%';

CREATE TABLE rekor.events (
    ts TIMESTAMP,
    idx BIGINT UNSIGNED,
    sub VARCHAR(255),
    pubkey_hash VARCHAR(70)
);

