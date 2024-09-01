CREATE TABLE user (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL DEFAULT '',
    email VARCHAR(100) NOT NULL DEFAULT '',
    password VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url VARCHAR(255) NOT NULL DEFAULT '',
    introduction VARCHAR(255) NOT NULL DEFAULT '',
    profile_info VARCHAR(255) NOT NULL DEFAULT '',
    reputation INT NOT NULL DEFAULT 0 COMMENT 'Reputation can be used to vote up or down',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1: active, 0: deleted',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY(id),
    UNIQUE KEY(username),
    UNIQUE KEY(email),
    INDEX idx_status (status),
    INDEX idx_updated_at (updated_at, status)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;