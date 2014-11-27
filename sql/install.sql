/*
Navicat MySQL Data Transfer

Source Server         : 10.180.120.63
Source Server Version : 50538
Source Host           : 10.180.120.63:3308
Source Database       : appmsgsrv

Target Server Type    : MYSQL
Target Server Version : 50538
File Encoding         : 65001

Date: 2014-11-27 09:11:07
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for api_call
-- ----------------------------
DROP TABLE IF EXISTS `api_call`;
CREATE TABLE `api_call` (
  `id` varchar(64) NOT NULL,
  `customer_id` varchar(64) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `caller_id` varchar(64) DEFAULT NULL COMMENT 'user_id 或 application_id',
  `type` varchar(64) DEFAULT NULL,
  `api_name` varchar(255) DEFAULT NULL,
  `count` int(11) DEFAULT NULL,
  `sharding` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for apns_token
-- ----------------------------
DROP TABLE IF EXISTS `apns_token`;
CREATE TABLE `apns_token` (
  `id` varchar(64) NOT NULL,
  `user_id` varchar(64) DEFAULT NULL,
  `device_id` varchar(64) DEFAULT NULL,
  `apns_token` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for application
-- ----------------------------
DROP TABLE IF EXISTS `application`;
CREATE TABLE `application` (
  `id` varchar(64) NOT NULL,
  `name` varchar(45) DEFAULT NULL,
  `token` varchar(45) DEFAULT NULL,
  `type` varchar(45) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `sort` int(11) DEFAULT NULL,
  `level` int(11) DEFAULT NULL,
  `avatar` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  `name_py` varchar(45) DEFAULT NULL,
  `name_quanpin` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for client
-- ----------------------------
DROP TABLE IF EXISTS `client`;
CREATE TABLE `client` (
  `id` varchar(64) NOT NULL,
  `user_id` varchar(64) DEFAULT NULL,
  `type` varchar(45) DEFAULT NULL,
  `device_id` varchar(255) DEFAULT NULL COMMENT '多个的话以 , 分隔',
  `latest_login_time` datetime DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for client_version
-- ----------------------------
DROP TABLE IF EXISTS `client_version`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for customer
-- ----------------------------
DROP TABLE IF EXISTS `customer`;
CREATE TABLE `customer` (
  `id` varchar(64) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for group_msg
-- ----------------------------
DROP TABLE IF EXISTS `group_msg`;
CREATE TABLE `group_msg` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) unsigned NOT NULL,
  `mid` bigint(20) unsigned NOT NULL,
  `ttl` bigint(20) NOT NULL,
  `msg` blob NOT NULL,
  `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `mtime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_group_msg_1` (`gid`,`mid`),
  KEY `ix_group_msg_1` (`ttl`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for opertion
-- ----------------------------
DROP TABLE IF EXISTS `opertion`;
CREATE TABLE `opertion` (
  `id` varchar(64) NOT NULL COMMENT 'ID',
  `app_id` varchar(64) DEFAULT NULL COMMENT '应用ID',
  `content` varchar(45) DEFAULT NULL COMMENT '操作项显示内容',
  `action` varchar(255) DEFAULT NULL COMMENT '操作项对应的URL',
  `msg_type` char(1) DEFAULT NULL COMMENT '1 返回页面信息  2 返回JSON串',
  `sort` int(11) DEFAULT NULL,
  `parent_id` varchar(64) DEFAULT NULL COMMENT '操作项的父操作项',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for org
-- ----------------------------
DROP TABLE IF EXISTS `org`;
CREATE TABLE `org` (
  `id` varchar(64) NOT NULL,
  `name` varchar(120) NOT NULL DEFAULT '' COMMENT '编码',
  `short_name` varchar(60) NOT NULL DEFAULT '',
  `parent_id` varchar(64) NOT NULL DEFAULT '',
  `location` varchar(256) NOT NULL DEFAULT '',
  `tenant_id` varchar(64) NOT NULL DEFAULT '',
  `sort` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for org_user
-- ----------------------------
DROP TABLE IF EXISTS `org_user`;
CREATE TABLE `org_user` (
  `id` varchar(64) NOT NULL,
  `org_id` varchar(64) NOT NULL DEFAULT '',
  `user_id` varchar(64) NOT NULL DEFAULT '',
  `sort` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for private_msg
-- ----------------------------
DROP TABLE IF EXISTS `private_msg`;
CREATE TABLE `private_msg` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `skey` varchar(64) NOT NULL,
  `mid` bigint(20) unsigned NOT NULL,
  `ttl` bigint(20) NOT NULL,
  `msg` blob NOT NULL,
  `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `mtime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_private_msg_1` (`skey`,`mid`),
  KEY `ix_private_msg_1` (`ttl`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for public_msg
-- ----------------------------
DROP TABLE IF EXISTS `public_msg`;
CREATE TABLE `public_msg` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `mid` bigint(20) unsigned NOT NULL,
  `ttl` bigint(20) NOT NULL,
  `msg` blob NOT NULL,
  `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `mtime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`id`),
  KEY `ix_public_msg_1` (`ttl`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for public_msg_log
-- ----------------------------
DROP TABLE IF EXISTS `public_msg_log`;
CREATE TABLE `public_msg_log` (
  `mid` bigint(20) unsigned NOT NULL,
  `stime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `ftime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  UNIQUE KEY `ux_public_msg_log_1` (`mid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for push_cnt
-- ----------------------------
DROP TABLE IF EXISTS `push_cnt`;
CREATE TABLE `push_cnt` (
  `id` varchar(64) NOT NULL,
  `customer_id` varchar(64) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `caller_id` varchar(64) DEFAULT NULL,
  `type` varchar(64) DEFAULT NULL,
  `push_type` varchar(64) DEFAULT NULL COMMENT 'qun/user',
  `count` int(11) DEFAULT NULL,
  `sharding` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for qun
-- ----------------------------
DROP TABLE IF EXISTS `qun`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='一个群（qun）包含了一个或多个组（group）';

-- ----------------------------
-- Table structure for qun_user
-- ----------------------------
DROP TABLE IF EXISTS `qun_user`;
CREATE TABLE `qun_user` (
  `id` varchar(64) NOT NULL,
  `qun_id` varchar(64) NOT NULL DEFAULT '',
  `user_id` varchar(64) NOT NULL DEFAULT '',
  `sort` int(11) NOT NULL DEFAULT '0',
  `role` int(11) NOT NULL DEFAULT '0' COMMENT '0：群主；1：管理员；2：成员',
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for quota
-- ----------------------------
DROP TABLE IF EXISTS `quota`;
CREATE TABLE `quota` (
  `id` varchar(64) NOT NULL,
  `customer_id` varchar(64) DEFAULT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `api_name` varchar(255) DEFAULT NULL,
  `type` varchar(45) DEFAULT NULL COMMENT '计数/到期时间',
  `value` varchar(45) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for resource
-- ----------------------------
DROP TABLE IF EXISTS `resource`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for session
-- ----------------------------
DROP TABLE IF EXISTS `session`;
CREATE TABLE `session` (
  `id` varchar(64) NOT NULL,
  `type` varchar(45) DEFAULT NULL,
  `user_id` varchar(64) DEFAULT NULL,
  `state` varchar(45) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for tenant
-- ----------------------------
DROP TABLE IF EXISTS `tenant`;
CREATE TABLE `tenant` (
  `id` varchar(64) NOT NULL,
  `code` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(128) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `customer_id` varchar(64) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for user_user
-- ----------------------------
DROP TABLE IF EXISTS `user_user`;
CREATE TABLE `user_user` (
  `id` varchar(64) NOT NULL,
  `from_user_id` varchar(64) DEFAULT NULL,
  `to_user_id` varchar(64) DEFAULT NULL,
  `remark_name` varchar(45) DEFAULT NULL,
  `sort` int(11) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
