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
  `domain_id` int NOT NULL,
  `domain_name` varchar(255) NOT NULL,
  `domain_url` varchar(255) NOT NULL,
  `user_id` int DEFAULT NULL,
  `php_version` varchar(255) DEFAULT '8.2'
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `plans`
--

CREATE TABLE `plans` (
  `id` int NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text,
  `domains_limit` int NOT NULL,
  `websites_limit` int NOT NULL,
  `disk_limit` text NOT NULL,
  `inodes_limit` bigint NOT NULL,
  `db_limit` int NOT NULL,
  `cpu` varchar(50) DEFAULT NULL,
  `ram` varchar(50) DEFAULT NULL,
  `docker_image` varchar(50) DEFAULT NULL,
  `bandwidth` int DEFAULT NULL,
  `storage_file` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


--
-- Dumping data for table `plans`
--

LOCK TABLES `plans` WRITE;
/*!40000 ALTER TABLE `plans` DISABLE KEYS */;
INSERT INTO `plans` VALUES (1,'ubuntu_nginx_mysql','Unlimited disk space and Nginx',0,10,'10 GB',1000000,0,'1','1g','openpanel/nginx',100,'0 GB'),(2,'ubuntu_apache_mysql','Unlimited disk space and Apache',0,10,'10 GB',1000000,0,'1','1g','openpanel/apache',100,'0 GB'),(3,'ubuntu_apache_mariadb','Unlimited disk space and Apache+MariaDB',0,10,'10 GB',1000000,0,'1','1g','openpanel/apache-mariadb',100,'0 GB'),(4,'ubuntu_nginx_mariadb','Unlimited disk space and Nginx+MariaDB',0,10,'10 GB',1000000,0,'1','1g','openpanel/nginx-mariadb',100,'0 GB');
/*!40000 ALTER TABLE `plans` ENABLE KEYS */;
UNLOCK TABLES;




-- --------------------------------------------------------

--
-- Table structure for table `sites`
--

CREATE TABLE `sites` (
  `id` int NOT NULL,
  `site_name` varchar(255) DEFAULT NULL,
  `domain_id` int DEFAULT NULL,
  `admin_email` varchar(255) DEFAULT NULL,
  `version` varchar(20) DEFAULT NULL,
  `created_date` datetime DEFAULT CURRENT_TIMESTAMP,
  `type` varchar(50) DEFAULT NULL,
  `ports` int DEFAULT NULL,
  `path` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` int NOT NULL,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `services` varchar(255) CHARACTER SET latin1 COLLATE latin1_swedish_ci DEFAULT '1,2,3,4,5,6,7,8,9,10,11,12',
  `user_domains` varchar(255) NOT NULL DEFAULT '',
  `twofa_enabled` tinyint(1) DEFAULT '0',
  `otp_secret` varchar(255) DEFAULT NULL,
  `plan` varchar(255) CHARACTER SET latin1 COLLATE latin1_swedish_ci DEFAULT NULL,
  `registered_date` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `plan_id` int DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

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
-- AUTO_INCREMENT for table `domains`
--
ALTER TABLE `domains`
  MODIFY `domain_id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;

--
-- AUTO_INCREMENT for table `plans`
--
ALTER TABLE `plans`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `sites`
--
ALTER TABLE `sites`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1;

--
-- Constraints for dumped tables
--

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

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
