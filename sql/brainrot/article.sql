CREATE TABLE article (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    author_id INT UNSIGNED NOT NULL,
    author VARCHAR(50) NOT NULL, -- The redundancy can help us reduce queries
    title VARCHAR(255) NOT NULL,
    content MEDIUMTEXT NOT NULL,
    tags VARCHAR(255) NOT NULL DEFAULT '',
    poster VARCHAR(255) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0: active, 1: deleted',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uk_author_title (author_id, title),
    INDEX idx_status (status),
    INDEX idx_author_status (author_id, status),
    INDEX idx_updated_at (updated_at, status)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;