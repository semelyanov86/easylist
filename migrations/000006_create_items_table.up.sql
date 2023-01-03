CREATE TABLE `items`
(
    `id`            BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `user_id`       BIGINT          NOT NULL REFERENCES users ON DELETE CASCADE,
    `list_id`       BIGINT          NOT NULL REFERENCES lists ON DELETE CASCADE,
    `name`          VARCHAR(255)    NOT NULL,
    `description`   TEXT,
    `quantity`      INT             NOT NULL DEFAULT 0,
    `quantity_type` VARCHAR(255)    NOT NULL DEFAULT 'piece',
    `price`         DECIMAL(8, 2)   NOT NULL COMMENT 'Цена объекта DECIMAL(10,2)',
    `is_starred`    BOOL      NOT NULL DEFAULT false,
    `file`          TEXT            NOT NULL COMMENT 'Изображение товара',
    `created_at`    DATETIME        NOT NULL DEFAULT NOW(),
    `updated_at`    DATETIME        NOT NULL DEFAULT NOW(),
    `version`       INT             NOT NULL DEFAULT 1,
    `order`         INT             NOT NULL DEFAULT 1
);