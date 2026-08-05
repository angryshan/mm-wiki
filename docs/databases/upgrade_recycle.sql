-- ================================================
-- MM-Wiki 回收站功能 数据库升级脚本
-- author: angryshan
-- ================================================

-- 1. mw_space 表新增：回收站保留天数
-- 默认30天，0天则为立即删除
ALTER TABLE `mw_space` ADD COLUMN `recycle_keep_days` int(11) NOT NULL DEFAULT '30' COMMENT '回收站保留天数，0为立即删除，默认30天';

-- 2. mw_document 表新增：删除时间
-- 0 表示未删除/正常文档，大于0表示进入回收站的时间戳
ALTER TABLE `mw_document` ADD COLUMN `deleted_time` int(11) NOT NULL DEFAULT '0' COMMENT '删除时间，0表示未删除，大于0表示进入回收站的时间';
