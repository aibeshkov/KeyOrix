-- Rollback RBAC Seed Data Migration

-- Remove default role permissions
DELETE FROM role_permissions WHERE role_id IN (
  SELECT id FROM roles WHERE name IN ('super_admin', 'admin', 'editor', 'viewer', 'auditor')
);

-- Remove default permissions
DELETE FROM permissions WHERE resource IN ('secrets', 'users', 'roles', 'system', 'audit', 'namespaces');

-- Remove default roles
DELETE FROM roles WHERE name IN ('super_admin', 'admin', 'editor', 'viewer', 'auditor');

-- Remove default environments
DELETE FROM environments WHERE name IN ('production', 'staging', 'development', 'testing');

-- Remove default zones
DELETE FROM zones WHERE name IN ('global', 'us-east-1', 'us-west-2', 'eu-west-1');

-- Remove default namespaces
DELETE FROM namespaces WHERE name IN ('default', 'production', 'staging', 'development');