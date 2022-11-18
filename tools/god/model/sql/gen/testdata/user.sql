CREATE TABLE `test_user` (
     `id` bigint(20) NOT NULL AUTO_INCREMENT,
     `mobile` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '手机号',
     `class` bigint(20) NOT NULL COMMENT '班级编号',
     `name` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '姓名',
     `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
     `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
     PRIMARY KEY (`id`),
     UNIQUE KEY `mobile_unique` (`mobile`),
     UNIQUE KEY `class_name_unique` (`class`,`name`),
     KEY `create_index` (`create_time`),
     KEY `name_index` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;