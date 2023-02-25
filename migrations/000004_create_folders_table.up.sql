CREATE TABLE IF NOT EXISTS `folders`
(
    `id`         BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT                      REFERENCES users ON DELETE CASCADE,
    `name`       VARCHAR(255)                NOT NULL COMMENT 'Наименование папки',
    `icon`       VARCHAR(255)                NOT NULL DEFAULT 'mdi-folder' COMMENT 'Иконка с папкой fontawesome.com',
    `version`    INT                         NOT NULL DEFAULT 1,
    `order`      INT                         NOT NULL DEFAULT 1 COMMENT 'По данному полю делаем сортировку',
    `created_at` DATETIME                    NOT NULL DEFAULT NOW(),
    `updated_at` DATETIME                    NOT NULL DEFAULT NOW()
);

INSERT INTO folders (name) values ('default');