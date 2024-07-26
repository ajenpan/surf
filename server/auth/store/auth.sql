create DATABASE if not exists auth default character set = 'utf8mb4';

use auth;

CREATE TABLE IF NOT EXISTS `users` (
  `uid` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'user unique id',
  `uname` varchar(64) CHARACTER SET utf8mb4 NOT NULL COMMENT 'user name',
  `passwd` varchar(64) NOT NULL DEFAULT '' COMMENT 'password',
  `nickname` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `avatar` varchar(1024) NOT NULL DEFAULT '',
  `gender` tinyint(4) NOT NULL DEFAULT 0,
  `phone` varchar(32) NOT NULL DEFAULT '',
  `email` varchar(64) NOT NULL DEFAULT '',
  `stat` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'user status code',
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '',
  `update_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `UQE_uname` (`uname`)
) ENGINE = InnoDB AUTO_INCREMENT = 100000 DEFAULT CHARSET = utf8mb4;

insert into  users (uname, passwd, nickname) values ('test', '123456', 'test');