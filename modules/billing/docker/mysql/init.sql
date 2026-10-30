CREATE DATABASE IF NOT EXISTS `billing` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'billing'@'%' IDENTIFIED BY 'billing';
GRANT ALL PRIVILEGES ON `billing`.* TO 'billing'@'%';
