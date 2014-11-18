delimiter $$

CREATE TABLE `api_call` (
  `id` varchar(64) NOT NULL,
  `customer_id` varchar(64) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `caller_id` varchar(64) DEFAULT NULL COMMENT 'user_id 或 application_id',
  `api_name` varchar(255) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `sharding` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$

delimiter $$

CREATE TABLE `quota` (
  `id` varchar(64) NOT NULL,
  `customer_id` varchar(64) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `api_name` varchar(255) DEFAULT NULL,
  `key` varchar(45) DEFAULT NULL COMMENT '计数/到期时间',
  `value` varchar(45) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$

delimiter $$

CREATE TABLE `application` (
  `id` varchar(64) NOT NULL,
  `name` varchar(45) DEFAULT NULL,
  `token` varchar(45) DEFAULT NULL,
  `type` varchar(45) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `sort` int(11) DEFAULT NULL,
  `level` int(11) DEFAULT NULL,
  `icon` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `session` (
  `id` varchar(64) NOT NULL,
  `type` varchar(45) DEFAULT NULL,
  `user_id` varchar(64) DEFAULT NULL,
  `state` varchar(45) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$

delimiter $$

CREATE TABLE `apns_token` (
  `id` varchar(64) NOT NULL,
  `user_id` varchar(64) DEFAULT NULL,
  `device_id` varchar(64) DEFAULT NULL,
  `apns_token` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$

delimiter $$

CREATE TABLE `customer` (
  `id` varchar(64) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$

delimiter $$

CREATE TABLE `resource` (
  `id` varchar(64) NOT NULL,
  `customer_id` varchar(64) DEFAULT NULL,
  `name` varchar(45) DEFAULT NULL,
  `description` varchar(512) DEFAULT NULL,
  `type` varchar(45) DEFAULT NULL,
  `content` varchar(255) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$

delimiter $$

CREATE TABLE `client` (
  `id` varchar(64) NOT NULL,
  `user_id` varchar(64) DEFAULT NULL,
  `type` varchar(45) DEFAULT NULL,
  `device_id` varchar(255) DEFAULT NULL COMMENT '多个的话以 , 分隔',
  `latest_login_time` datetime DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `client_version` (
  `id` varchar(64) NOT NULL,
  `type` varchar(45) DEFAULT NULL,
  `ver_code` int(11) DEFAULT NULL,
  `ver_name` varchar(45) DEFAULT NULL,
  `ver_description` text,
  `download_url` varchar(255) DEFAULT NULL,
  `file_name` varchar(45) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `org` (
  `id` varchar(64) NOT NULL,
  `name` varchar(120) NOT NULL DEFAULT '' COMMENT '编码',
  `short_name` varchar(60) NOT NULL DEFAULT '',
  `parent_id` varchar(64) NOT NULL DEFAULT '',
  `location` varchar(256) NOT NULL DEFAULT '',
  `tenant_id` varchar(64) NOT NULL DEFAULT '',
  `sort` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `org_user` (
  `id` varchar(64) NOT NULL,
  `org_id` varchar(64) NOT NULL DEFAULT '',
  `user_id` varchar(64) NOT NULL DEFAULT '',
  `sort` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `qun` (
  `id` varchar(64) NOT NULL,
  `creator_id` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(45) NOT NULL DEFAULT '',
  `description` varchar(255) NOT NULL DEFAULT '',
  `max_member` int(11) NOT NULL DEFAULT '0',
  `avatar` varchar(255) NOT NULL DEFAULT '',
  `tenant_id` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `qun_user` (
  `id` varchar(64) NOT NULL,
  `qun_id` varchar(64) NOT NULL DEFAULT '',
  `user_id` varchar(64) NOT NULL DEFAULT '',
  `sort` int(11) NOT NULL DEFAULT '0',
  `role` int(11) NOT NULL DEFAULT '0' COMMENT '0：群主；1：管理员；2：成员',
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `tenant` (
  `id` varchar(64) NOT NULL,
  `code` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(128) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `customer_id` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `user` (
  `id` varchar(64) NOT NULL,
  `name` varchar(45) NOT NULL DEFAULT '',
  `nickname` varchar(45) NOT NULL DEFAULT '',
  `avatar` varchar(45) NOT NULL DEFAULT '',
  `name_py` varchar(45) NOT NULL DEFAULT '',
  `name_quanpin` varchar(255) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `rand` int(11) NOT NULL DEFAULT '0',
  `password` varchar(32) NOT NULL DEFAULT '',
  `tenant_id` varchar(64) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT '0',
  `email` varchar(64) NOT NULL DEFAULT '',
  `mobile` varchar(11) NOT NULL DEFAULT '',
  `area` varchar(128) NOT NULL DEFAULT '',
  `created` datetime NOT NULL,
  `updated` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


delimiter $$

CREATE TABLE `user_user` (
  `id` varchar(64) NOT NULL,
  `from_user_id` varchar(64) DEFAULT NULL,
  `to_user_id` varchar(64) DEFAULT NULL,
  `remark_name` varchar(45) DEFAULT NULL,
  `sort` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8$$


