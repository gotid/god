CREATE TABLE `shorturl` (
    `shorten` varchar(255) NOT NULL COMMENT 'shorten key',
    `url` varchar(255) NOT NULL DEFAULT '' COMMENT 'original url',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`shorten`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;