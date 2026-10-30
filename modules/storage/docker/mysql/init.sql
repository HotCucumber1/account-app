CREATE DATABASE IF NOT EXISTS `storage` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'storage'@'%' IDENTIFIED BY 'storage';
GRANT ALL PRIVILEGES ON `storage`.* TO 'storage'@'%';
