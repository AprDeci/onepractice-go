CREATE TABLE IF NOT EXISTS `user_points` (
  `user_id` bigint NOT NULL,
  `balance` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `point_transactions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `delta` bigint NOT NULL DEFAULT 0,
  `balance_after` bigint NOT NULL DEFAULT 0,
  `type` varchar(32) NOT NULL DEFAULT '',
  `biz_id` varchar(64) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 0,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pt_idem` (`user_id`, `type`, `biz_id`),
  KEY `idx_pt_user_created` (`user_id`, `created_at`),
  KEY `idx_pt_sweep` (`status`, `type`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `point_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `action` varchar(32) NOT NULL DEFAULT '',
  `value` bigint NOT NULL DEFAULT 0,
  `effective_from` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `effective_to` datetime(3) DEFAULT NULL,
  `enabled` tinyint NOT NULL DEFAULT 1,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_rule_action_effective` (`action`, `enabled`, `effective_from`, `effective_to`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
