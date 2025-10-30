-- MySQL dump 10.13  Distrib 8.0.36, for Linux (x86_64)
--
-- Host: localhost    Database: panel
-- ------------------------------------------------------
-- Server version	8.0.36-0ubuntu0.22.04.1

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `panel`
--

-- --------------------------------------------------------

--
-- Table structure for table `domains`
--

CREATE TABLE `domains` (
  `domain_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `docroot` VARCHAR(255) NOT NULL,
  `domain_url` VARCHAR(255) NOT NULL,
  `user_id` INT UNSIGNED DEFAULT NULL,
  `php_version` VARCHAR(255) DEFAULT '8.2'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `plans`
--

CREATE TABLE `plans` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `domains_limit` INT NOT NULL,
  `websites_limit` INT NOT NULL,
  `email_limit` INT NOT NULL DEFAULT 0,
  `ftp_limit` INT NOT NULL DEFAULT 0,
  `disk_limit` TEXT NOT NULL,
  `inodes_limit` BIGINT NOT NULL,
  `db_limit` INT NOT NULL,
  `cpu` VARCHAR(50) DEFAULT NULL,
  `ram` VARCHAR(50) DEFAULT NULL,
  `bandwidth` INT DEFAULT NULL,
  `feature_set` VARCHAR(255) DEFAULT 'default'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


--
-- Dumping data for table `plans`
--

LOCK TABLES `plans` WRITE;
/*!40000 ALTER TABLE `plans` DISABLE KEYS */;
INSERT INTO `plans` (id, name, description, domains_limit, websites_limit, email_limit, ftp_limit, disk_limit, inodes_limit, db_limit, cpu, ram, bandwidth, feature_set) 
VALUES
(1, 'Standard plan', 'Small plan for testing', 0, 10, 0, 0, '5 GB', 1000000, 0, '2', '2g', 10, 'basic'),
(2, 'Developer Plus', '4 cores, 6G ram', 0, 10, 0, 0, '10 GB', 1000000, 0, '4', '6g', 100, 'default');
/*!40000 ALTER TABLE `plans` ENABLE KEYS */;
UNLOCK TABLES;




-- --------------------------------------------------------

--
-- Table structure for table `sites`
--

CREATE TABLE `sites` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `site_name` VARCHAR(255) DEFAULT NULL,
  `domain_id` INT UNSIGNED DEFAULT NULL,
  `admin_email` VARCHAR(255) DEFAULT NULL,
  `version` VARCHAR(20) DEFAULT NULL,
  `created_date` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `type` VARCHAR(50) DEFAULT NULL,
  `container` VARCHAR(255) DEFAULT NULL,
  `ports` VARCHAR(255) DEFAULT NULL,
  `path` VARCHAR(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(255) NOT NULL,
  `password` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  `owner` VARCHAR(255) CHARACTER SET latin1 COLLATE latin1_swedish_ci DEFAULT NULL,
  `user_domains` INT NOT NULL DEFAULT 0,
  `twofa_enabled` TINYINT(1) DEFAULT 0,
  `otp_secret` VARCHAR(255) DEFAULT NULL,
  `plan` VARCHAR(255) CHARACTER SET latin1 COLLATE latin1_swedish_ci DEFAULT NULL,
  `registered_date` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  `server` VARCHAR(255) DEFAULT 'default',
  `plan_id` INT UNSIGNED DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



--
-- Table structure for table `active_sessions`
--

CREATE TABLE `active_sessions` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` INT UNSIGNED NOT NULL,
  `session_token` VARCHAR(255) NOT NULL,
  `ip_address` VARCHAR(45) DEFAULT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `expires_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


--
-- Indexes for dumped tables
--

--
-- Indexes for table `domains`
--
ALTER TABLE `domains`
  ADD PRIMARY KEY (`domain_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `plans`
--
ALTER TABLE `plans`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sites`
--
ALTER TABLE `sites`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_sites_domain` (`domain_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_plan_id` (`plan_id`);

--
-- AUTO_INCREMENT for dumped tables
--


--
-- Constraints for dumped tables
--


--
-- Constraints for table `active_sessions`
--
ALTER TABLE `active_sessions`
  ADD CONSTRAINT `active_sessions_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);


--
-- Constraints for table `domains`
--
ALTER TABLE `domains`
  ADD CONSTRAINT `domains_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

--
-- Constraints for table `sites`
--
ALTER TABLE `sites`
  ADD CONSTRAINT `fk_sites_domain` FOREIGN KEY (`domain_id`) REFERENCES `domains` (`domain_id`);

--
-- Constraints for table `users`
--
ALTER TABLE `users`
  ADD CONSTRAINT `fk_plan_id` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON DELETE SET NULL;
COMMIT;




--
-- View for calculating number of user domains
--

CREATE TRIGGER increment_user_domains
AFTER INSERT ON domains
FOR EACH ROW
UPDATE users SET user_domains = user_domains + 1 WHERE id = NEW.user_id;

CREATE TRIGGER decrement_user_domains
AFTER DELETE ON domains
FOR EACH ROW
UPDATE users SET user_domains = user_domains - 1 WHERE id = OLD.user_id;


/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
