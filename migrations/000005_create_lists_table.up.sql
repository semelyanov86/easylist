CREATE TABLE IF NOT EXISTS `lists`
(
    `id`         BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT          NOT NULL REFERENCES users ON DELETE CASCADE,
    `folder_id`  BIGINT          NOT NULL REFERENCES folders ON DELETE CASCADE,
    `name`       VARCHAR(255)    NOT NULL,
    `icon`       VARCHAR(255)    NOT NULL DEFAULT 'mdi-view-list',
    `version`    INT             NOT NULL DEFAULT 1,
    `order`      INT             NOT NULL DEFAULT 1,
    `link`       CHAR(36)        COMMENT 'Уникальный идентификатор для публичной ссылки',
    `created_at` DATETIME        NOT NULL DEFAULT NOW(),
    `updated_at` DATETIME        NOT NULL DEFAULT NOW()
);