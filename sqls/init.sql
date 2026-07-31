-- 用户管理系统数据库初始化脚本
-- 创建时间: 2026-01-29
-- 数据库: MySQL

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `landc_user_manager` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `landc_user_manager`;

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `email` varchar(100) NOT NULL COMMENT '邮箱',
  `password` varchar(100) NOT NULL COMMENT '密码（加密）',
  `nickname` varchar(50) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1-活跃, 0-禁用',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 登录历史表
CREATE TABLE IF NOT EXISTS `login_histories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '登录历史ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `ip` varchar(50) DEFAULT NULL COMMENT 'IP地址',
  `device_fingerprint` varchar(255) DEFAULT NULL COMMENT '设备指纹',
  `user_agent` varchar(255) DEFAULT NULL COMMENT '用户代理',
  `login_time` datetime(3) DEFAULT NULL COMMENT '登录时间',
  `status` tinyint DEFAULT NULL COMMENT '状态: 1-成功, 0-失败',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `device` varchar(100) DEFAULT NULL COMMENT '设备',
  `location` varchar(100) DEFAULT NULL COMMENT '位置',
  PRIMARY KEY (`id`),
  KEY `idx_login_histories_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录历史表';

-- 第三方用户元数据表
CREATE TABLE IF NOT EXISTS `third_party_user_metas` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '第三方用户ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `provider` varchar(50) NOT NULL COMMENT '第三方提供商（wechat/dingtalk/github/alipay）',
  `open_id` varchar(100) NOT NULL COMMENT 'OpenID',
  `union_id` varchar(100) DEFAULT NULL COMMENT 'UnionID',
  `access_token` varchar(500) DEFAULT NULL COMMENT '访问令牌（AES-GCM 加密存储，base64 + enc: 前缀）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_third_party_user_metas_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方用户元数据表';

-- 信任设备表
CREATE TABLE IF NOT EXISTS `trusted_devices` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '信任设备ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `device_fingerprint` varchar(255) NOT NULL COMMENT '设备指纹',
  `device_name` varchar(100) DEFAULT NULL COMMENT '设备名称',
  `last_login_ip` varchar(50) DEFAULT NULL COMMENT '最后登录IP',
  `first_login_time` datetime(3) DEFAULT NULL COMMENT '首次登录时间',
  `last_login_time` datetime(3) DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `last_login_location` varchar(100) DEFAULT NULL COMMENT '最后登录位置',
  PRIMARY KEY (`id`),
  KEY `idx_trusted_devices_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信任设备表';

-- 生物识别表
CREATE TABLE IF NOT EXISTS `biometrics` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '生物识别ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `biometric_type` varchar(20) DEFAULT NULL COMMENT '生物识别类型（fingerprint/face/iris）',
  `biometric_data` varchar(255) DEFAULT NULL COMMENT '生物识别数据',
  PRIMARY KEY (`id`),
  KEY `idx_biometrics_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生物识别表';
