-- username 的语义一直是昵称：不唯一、不参与登录；账号唯一标识是 email。
-- 本迁移只改列名与该列上的唯一索引，不动任何数据内容（RENAME COLUMN 保留原值）。
-- 幂等，可重复执行。

-- 1) username -> nickname（已改过则跳过）
SET @has_old := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'user' AND column_name = 'username');
SET @has_new := (SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'user' AND column_name = 'nickname');
SET @stmt := IF(@has_old = 1 AND @has_new = 0,
  'ALTER TABLE `user` RENAME COLUMN `username` TO `nickname`', 'SELECT 1');
PREPARE s FROM @stmt;
EXECUTE s;
DEALLOCATE PREPARE s;

-- 2) 昵称不再唯一：删掉 001_schema.sql 建的 uk_user_username（不存在则跳过）
SET @has_uk := (SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE() AND table_name = 'user' AND index_name = 'uk_user_username');
SET @stmt := IF(@has_uk > 0,
  'ALTER TABLE `user` DROP INDEX `uk_user_username`', 'SELECT 1');
PREPARE s FROM @stmt;
EXECUTE s;
DEALLOCATE PREPARE s;
