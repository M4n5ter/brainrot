CREATE TABLE comment (
    id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    article_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    commenter VARCHAR(50) NOT NULL,
    useful_count INT UNSIGNED NOT NULL DEFAULT 0,
    useless_count INT UNSIGNED NOT NULL DEFAULT 0,
    voter_ids TEXT NOT NULL,
    content TEXT NOT NULL,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1: active, 0: deleted',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY(id),
    INDEX idx_status (status),
    INDEX idx_article_id_status (article_id, status),
    INDEX idx_user_id_status (user_id, status),
    INDEX idx_commenter_status (commenter, status),
    INDEX idx_updated_at (updated_at, status)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;