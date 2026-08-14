-- ================================================
-- MM-Wiki 回收站菜单权限 数据库升级脚本
-- author: angryshan
-- ================================================

-- 插入回收站顶级菜单（和空间管理平级）
INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('回收站', 1, 'menu', '', '', 'glyphicon-trash', '', 1, 10, unix_timestamp(now()), unix_timestamp(now()));

-- 回收站列表页（显示在菜单上）
INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('回收站列表', (SELECT privilege_id FROM (SELECT privilege_id FROM mw_privilege WHERE name='回收站' AND type='menu' LIMIT 1) t), 'controller', 'recycle', 'list', 'glyphicon-list', '', 1, 101, unix_timestamp(now()), unix_timestamp(now()));

-- 回收站操作权限（不显示在菜单上）
INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('恢复文档', (SELECT privilege_id FROM (SELECT privilege_id FROM mw_privilege WHERE name='回收站' AND type='menu' LIMIT 1) t), 'controller', 'recycle', 'recover', 'glyphicon-list', '', 0, 102, unix_timestamp(now()), unix_timestamp(now()));

INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('彻底删除文档', (SELECT privilege_id FROM (SELECT privilege_id FROM mw_privilege WHERE name='回收站' AND type='menu' LIMIT 1) t), 'controller', 'recycle', 'remove', 'glyphicon-list', '', 0, 103, unix_timestamp(now()), unix_timestamp(now()));

INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('清空回收站', (SELECT privilege_id FROM (SELECT privilege_id FROM mw_privilege WHERE name='回收站' AND type='menu' LIMIT 1) t), 'controller', 'recycle', 'clear', 'glyphicon-list', '', 0, 104, unix_timestamp(now()), unix_timestamp(now()));

INSERT INTO mw_privilege (name, parent_id, type, controller, action, icon, target, is_display, sequence, create_time, update_time)
VALUES ('查看回收站文档', (SELECT privilege_id FROM (SELECT privilege_id FROM mw_privilege WHERE name='回收站' AND type='menu' LIMIT 1) t), 'controller', 'recycle', 'view', '', '', 0, 105, unix_timestamp(now()), unix_timestamp(now()));

-- 给管理员角色（role_id=2）分配回收站权限
INSERT INTO mw_role_privilege (role_id, privilege_id, create_time)
SELECT 2, privilege_id, unix_timestamp(now()) FROM mw_privilege WHERE controller='recycle';
