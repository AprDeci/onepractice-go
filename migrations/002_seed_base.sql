INSERT INTO `tb_book` (`bookid`, `bookname`, `voccount`, `status`) VALUES
  (1, 'CET-4', 0, 1),
  (2, 'CET-6', 0, 1)
ON DUPLICATE KEY UPDATE
  `bookname` = VALUES(`bookname`),
  `status` = VALUES(`status`);
