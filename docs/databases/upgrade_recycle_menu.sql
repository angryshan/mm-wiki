-- ================================================
-- MM-Wiki 回收站菜单权限 数据库升级脚本
-- author: angryshan
-- ================================================

-- 插入回收站菜单权限（放在空间管理下）
INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('回收站', 37, 'controller', 'recycle', 'list', 'glyphicon-recycle', '', 1, 511, unix_timestamp(now()), unix_timestamp(now()));

-- 回收站操作权限（不显示在菜单上）
INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('恢复文档', 37, 'controller', 'recycle', 'recover', 'glyphicon-list', '', 0, 512, unix_timestamp(now()), unix_timestamp(now()));

INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('彻底删除文档', 37, 'controller', 'recycle', 'remove', 'glyphicon-list', '', 0, 513, unix_timestamp(now()), unix_timestamp(now()));

INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('清空回收站', 37, 'controller', 'recycle', 'clear', 'glyphicon-list', '', 0, 514, unix_timestamp(now()), unix_timestamp(now()));

-- 给管理员角色（role_id=2）分配回收站权限
INSERT INTO mw_role_privilege (role_id, privilege_id, create_time)
SELECT 2, privilege_id, unix_timestamp(now()) FROM mw_privilege WHERE controller='recycle';
