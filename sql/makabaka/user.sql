CREATE TABLE user (
    id INT AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL DEFAULT '',
    email VARCHAR(100) NOT NULL DEFAULT '',
    password VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url VARCHAR(255) NOT NULL DEFAULT '',
    introduction VARCHAR(255) NOT NULL DEFAULT '',
    profile_info JSON NOT NULL, -- JSON 不支持默认值，但是用户一定会有一条 joined info
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY(id),
    UNIQUE KEY(username),
    UNIQUE KEY(email)
);