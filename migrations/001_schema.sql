CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL,
  `password` varchar(255) NOT NULL,
  `email` varchar(512) NOT NULL,
  `user_type` int NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_username` (`username`),
  UNIQUE KEY `uk_user_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `papers` (
  `paper_id` int NOT NULL,
  `paper_name` varchar(255) NOT NULL,
  `exam_year` int NOT NULL,
  `exam_month` int NOT NULL,
  `version` int NOT NULL DEFAULT 1,
  `total_time` int NOT NULL DEFAULT 0,
  `type` varchar(32) NOT NULL,
  PRIMARY KEY (`paper_id`),
  KEY `idx_papers_type_year` (`type`, `exam_year`, `exam_month`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `questions` (
  `question_id` int NOT NULL,
  `paper_id` int NOT NULL,
  `part_name` varchar(128) NOT NULL DEFAULT '',
  `section_name` varchar(128) NOT NULL DEFAULT '',
  `question_type` varchar(64) NOT NULL DEFAULT '',
  `question_order` int NOT NULL DEFAULT 0,
  `content` longtext,
  `correct_answer` json,
  `reading_split_question` json,
  `options` json,
  `word_bank` json,
  `matching_data` json,
  `listenurl` varchar(1024) NOT NULL DEFAULT '',
  PRIMARY KEY (`question_id`),
  KEY `idx_questions_paper_order` (`paper_id`, `question_order`),
  KEY `idx_questions_paper_type` (`paper_id`, `question_type`),
  CONSTRAINT `fk_questions_paper` FOREIGN KEY (`paper_id`) REFERENCES `papers` (`paper_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `paper_rate_mapping` (
  `paperId` int NOT NULL,
  `rating` int NOT NULL DEFAULT 0,
  `number` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`paperId`),
  CONSTRAINT `fk_paper_rate_mapping_paper` FOREIGN KEY (`paperId`) REFERENCES `papers` (`paper_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `tb_vocabulary` (
  `wordid` int unsigned NOT NULL AUTO_INCREMENT,
  `spelling` varchar(255) NOT NULL,
  `UKphonetic` varchar(255) NOT NULL DEFAULT '',
  `USphonetic` varchar(255) NOT NULL DEFAULT '',
  `paraphrase` text,
  `frequency` double NOT NULL DEFAULT 0,
  PRIMARY KEY (`wordid`),
  UNIQUE KEY `uk_tb_vocabulary_spelling` (`spelling`),
  KEY `idx_tb_vocabulary_frequency` (`frequency`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `tb_book` (
  `bookid` int unsigned NOT NULL AUTO_INCREMENT,
  `bookname` varchar(255) NOT NULL,
  `voccount` int DEFAULT NULL,
  `status` int DEFAULT NULL,
  PRIMARY KEY (`bookid`),
  UNIQUE KEY `uk_tb_book_bookname` (`bookname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `tb_voc_book` (
  `vocbkid` int NOT NULL AUTO_INCREMENT,
  `wordid` int unsigned DEFAULT NULL,
  `bookid` int unsigned DEFAULT NULL,
  PRIMARY KEY (`vocbkid`),
  UNIQUE KEY `uk_tb_voc_book_word_book` (`wordid`, `bookid`),
  KEY `idx_tb_voc_book_bookid` (`bookid`),
  CONSTRAINT `fk_tb_voc_book_word` FOREIGN KEY (`wordid`) REFERENCES `tb_vocabulary` (`wordid`) ON DELETE CASCADE,
  CONSTRAINT `fk_tb_voc_book_book` FOREIGN KEY (`bookid`) REFERENCES `tb_book` (`bookid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `tb_voc_examples` (
  `exapid` int NOT NULL AUTO_INCREMENT,
  `wordid` int unsigned DEFAULT NULL,
  `en` text,
  `cn` text,
  `heat` int DEFAULT NULL,
  `adddate` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`exapid`),
  KEY `idx_tb_voc_examples_wordid` (`wordid`),
  CONSTRAINT `fk_tb_voc_examples_word` FOREIGN KEY (`wordid`) REFERENCES `tb_vocabulary` (`wordid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `user_word_favorites` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `wordid` int unsigned NOT NULL,
  `paper_id` int DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_word_favorites_user_word` (`user_id`, `wordid`),
  KEY `idx_user_word_favorites_wordid` (`wordid`),
  KEY `idx_user_word_favorites_paper_id` (`paper_id`),
  CONSTRAINT `fk_user_word_favorites_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_word_favorites_word` FOREIGN KEY (`wordid`) REFERENCES `tb_vocabulary` (`wordid`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_word_favorites_paper` FOREIGN KEY (`paper_id`) REFERENCES `papers` (`paper_id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
