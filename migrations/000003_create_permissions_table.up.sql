CREATE TABLE IF NOT EXISTS `permissions`
(
    `id`   BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `code` VARCHAR(255)    NOT NULL COMMENT 'Наименование разрешения'
);
CREATE TABLE IF NOT EXISTS `users_permissions`
(
    `id`            BIGINT UNSIGNED PRIMARY KEY NOT NULL AUTO_INCREMENT,
    `user_id`       BIGINT          NOT NULL REFERENCES users ON DELETE CASCADE,
    `permission_id` BIGINT          NOT NULL REFERENCES permissions ON DELETE CASCADE,
    `created_at`    DATETIME        NOT NULL DEFAULT NOW()
);

-- Add necessary permissions to the table.
INSERT INTO permissions (code)
VALUES
    ('folders:read'),
    ('folders:write'),
    ('lists:write'),
    ('lists:read'),
    ('items:read'),
    ('items:write');