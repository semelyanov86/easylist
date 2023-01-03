CREATE TABLE IF NOT EXISTS `tokens`
(
    `id`         BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `hash`       BINARY(16)      NOT NULL COMMENT 'Хэш, зашифрованный по алгоритму sha-256, длиной 32 или 128',
    `user_id`    BIGINT          NOT NULL REFERENCES users ON DELETE CASCADE,
    `expired_at` DATETIME        NOT NULL COMMENT 'Дата истечения срока действия',
    `scope`      VARCHAR(255)    NOT NULL COMMENT 'Тип токена'
);
