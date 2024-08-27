CREATE TABLE user (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL DEFAULT '',
    email VARCHAR(100) NOT NULL DEFAULT '',
    password VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url VARCHAR(255) NOT NULL DEFAULT '',
    introduction VARCHAR(255) NOT NULL DEFAULT '',
    profile_info VARCHAR(255) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0: active, 1: deleted',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY(id),
    UNIQUE KEY(username),
    UNIQUE KEY(email)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;