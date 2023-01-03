CREATE TABLE IF NOT EXISTS `users`
(
    `id`         BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT COMMENT 'Уникальный идентификатор пользователя',
    `name`       VARCHAR(255)    NOT NULL COMMENT 'Имя пользователя для отображения на сайте',
    `email`      VARCHAR(255)    UNIQUE NOT NULL COMMENT 'Уникальный адрес электронной почты пользователя',
    `password`   CHAR(255)       NOT NULL COMMENT 'Хранится хэш пароля длиной 60 символов',
    `created_at` DATETIME        NOT NULL DEFAULT NOW() COMMENT 'Дата создания',
    `updated_at` DATETIME        NOT NULL DEFAULT NOW() COMMENT 'Дата обновления',
    `is_active`  bool      NOT NULL DEFAULT false COMMENT 'Активирован ли пользователь или нет',
    `version`    INT             NOT NULL DEFAULT 1 COMMENT 'Версия записи, каждое обновление увеличивает индекс на 1'
);

