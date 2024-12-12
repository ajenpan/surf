CREATE TABLE `prop_detail` (
	`prop_id` INT NOT NULL COMMENT '道具',
	`prop_type` INT NOT NULL COMMENT '道具类型',
	`prop_group` INT NOT NULL COMMENT '道具组',
	`prop_name` VARCHAR ( 64 ) NOT NULL COMMENT '道具名称',
	`prop_desc` VARCHAR ( 1024 ) NOT NULL COMMENT '道具描述',
	`prop_icon` VARCHAR ( 128 ) NOT NULL COMMENT '道具图标',
	PRIMARY KEY ( `prop_id` ) USING BTREE 
) ENGINE = INNODB CHARACTER 
SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

CREATE TABLE `user_prop` (
	`uid` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户id',
	`prop_id` INT NOT NULL COMMENT '道具',
	`prop_cnt` BIGINT NOT NULL COMMENT '道具数量',
	`update_at` datetime NOT NULL COMMENT '最后更新时间',
	PRIMARY KEY ( `uid`, `prop_id` ) USING BTREE 
) ENGINE = INNODB CHARACTER 
SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- CREATE TABLE `user_num_mate`  (
--   `uid` bigint UNSIGNED NOT NULL AUTO_INCREMENT,  
--   `mate_key` varchar(64) NOT NULL COMMENT 'metakey',
--   `mate_value` bigint NOT NULL COMMENT 'metavalue',  
--   `update_at` datetime NOT NULL,
--   PRIMARY KEY (`uid`, `mate_key`) USING BTREE
-- ) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- CREATE TABLE `user_str_mate`  (
--   `uid` bigint UNSIGNED NOT NULL AUTO_INCREMENT,  
--   `mate_key` varchar(64) NOT NULL COMMENT 'metakey',
--   `mate_value` text NOT NULL COMMENT 'metavalue',  
--   `update_at` datetime NOT NULL,
--   PRIMARY KEY (`uid`, `mate_key`) USING BTREE
-- ) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;


