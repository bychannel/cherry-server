-- 开发帐号表
DROP TABLE IF EXISTS `account`;
CREATE TABLE `account` (
  `account_id` bigint(20) NOT NULL COMMENT '帐号id',
  `account_name` varchar(64) NOT NULL COMMENT '帐号名',
  `password` varchar(128) NOT NULL COMMENT '密码',
  `create_ip` varchar(64) DEFAULT NULL COMMENT '创建ip',
  `create_time` bigint(20) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`account_id`),
  UNIQUE KEY `account_name_key` (`account_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 用户绑定表
DROP TABLE IF EXISTS `user_bind`;
CREATE TABLE `user_bind` (
  `uid` bigint(20) NOT NULL COMMENT '用户唯一id',
  `sdk_id` int(11) DEFAULT '0' COMMENT 'sdk配置id',
  `pid` int(11) DEFAULT NULL COMMENT '平台id',
  `open_id` varchar(64) DEFAULT NULL COMMENT '平台帐号open_id',
  `bind_time` bigint(20) DEFAULT NULL COMMENT '绑定时间',
  `up_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后一次更新时间',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `pid_open_id_key` (`pid`,`open_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
