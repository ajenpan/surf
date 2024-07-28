CREATE TABLE `mail_list`  (
  `mailid` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '邮件id',
  `content` json NOT NULL COMMENT '内容',
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_by` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '创建人',
  `status` int NOT NULL DEFAULT 0 COMMENT '状态 0:正常, 1:失效',
  PRIMARY KEY (`mailid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

CREATE TABLE `mail_recv`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,  
  `numid` int NOT NULL,
  `mailid` int UNSIGNED NOT NULL,
  `mark` int UNSIGNED NOT NULL DEFAULT 0,
  `recv_at` datetime NOT NULL,
  `status` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `areaid_numid_mailid`(`areaid`, `numid`, `mailid`) USING BTREE,
  INDEX `mailid`(`mailid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;
