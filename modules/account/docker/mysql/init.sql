CREATE DATABASE IF NOT EXISTS `account` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'account'@'%' IDENTIFIED BY 'account';
GRANT ALL PRIVILEGES ON `account`.* TO 'account'@'%';
