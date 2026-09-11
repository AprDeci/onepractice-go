-- 作文标题实际存储的是题目全文（可能远超 255 字符），将 title 扩展为 TEXT。
-- 已在 003_grading.sql 中同步为 TEXT，此脚本用于修正已建库的旧表。
ALTER TABLE `essay_grading_results`
  MODIFY COLUMN `title` text NOT NULL;
