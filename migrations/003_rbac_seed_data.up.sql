-- RBAC Seed Data Migration
-- This migration adds default roles and permissions

-- Insert default roles
INSERT OR IGNORE INTO roles (name, description) VALUES 
  ('super_admin', 'Super Administrator with full system access'),
  ('admin', 'Administrator with full access to assigned namespaces'),
  ('editor', 'Can create, read, update secrets in assigned namespaces'),
  ('viewer', 'Read-only access to secrets in assigned namespaces'),
  ('auditor', 'Can view audit logs and system information');

-- Insert default permissions
INSERT OR IGNORE INTO permissions (name, description, resource, action) VALUES 
  -- Secret permissions
  ('secrets.read', 'Read secrets', 'secrets', 'read'),
  ('secrets.write', 'Create and update secrets', 'secrets', 'write'),
  ('secrets.delete', 'Delete secrets', 'secrets', 'delete'),
  ('secrets.admin', 'Full administrative access to secrets', 'secrets', 'admin'),
  
  -- User management permissions
  ('users.read', 'View user information', 'users', 'read'),
  ('users.write', 'Create and update users', 'users', 'write'),
  ('users.delete', 'Delete users', 'users', 'delete'),
  ('users.admin', 'Full administrative access to users', 'users', 'admin'),
  
  -- Role management permissions
  ('roles.read', 'View roles', 'roles', 'read'),
  ('roles.write', 'Create and update roles', 'roles', 'write'),
  ('roles.delete', 'Delete roles', 'roles', 'delete'),
  ('roles.admin', 'Full administrative access to roles', 'roles', 'admin'),
  ('roles.assign', 'Assign and remove roles from users', 'roles', 'assign'),
  
  -- System permissions
  ('system.read', 'View system information', 'system', 'read'),
  ('system.write', 'Modify system settings', 'system', 'write'),
  ('system.admin', 'Full administrative access to system', 'system', 'admin'),
  
  -- Audit permissions
  ('audit.read', 'View audit logs', 'audit', 'read'),
  ('audit.admin', 'Full administrative access to audit system', 'audit', 'admin'),
  
  -- Namespace permissions
  ('namespaces.read', 'View namespaces', 'namespaces', 'read'),
  ('namespaces.write', 'Create and update namespaces', 'namespaces', 'write'),
  ('namespaces.delete', 'Delete namespaces', 'namespaces', 'delete'),
  ('namespaces.admin', 'Full administrative access to namespaces', 'namespaces', 'admin');

-- Assign permissions to roles
-- Super Admin gets all permissions
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'super_admin';

-- Admin gets most permissions except super admin level
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'admin' 
  AND p.name IN (
    'secrets.read', 'secrets.write', 'secrets.delete',
    'users.read', 'users.write',
    'roles.read', 'roles.assign',
    'system.read',
    'audit.read',
    'namespaces.read', 'namespaces.write'
  );

-- Editor gets read/write access to secrets and basic user info
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'editor' 
  AND p.name IN (
    'secrets.read', 'secrets.write',
    'users.read',
    'namespaces.read'
  );

-- Viewer gets read-only access
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'viewer' 
  AND p.name IN (
    'secrets.read',
    'users.read',
    'namespaces.read'
  );

-- Auditor gets audit and system read access
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'auditor' 
  AND p.name IN (
    'audit.read', 'audit.admin',
    'system.read',
    'users.read',
    'roles.read',
    'namespaces.read'
  );

-- Insert default namespaces if they don't exist
INSERT OR IGNORE INTO namespaces (name, description) VALUES 
  ('default', 'Default namespace for general use'),
  ('production', 'Production environment namespace'),
  ('staging', 'Staging environment namespace'),
  ('development', 'Development environment namespace');

-- Insert default zones if they don't exist
INSERT OR IGNORE INTO zones (name, description) VALUES 
  ('global', 'Global zone for cross-region resources'),
  ('us-east-1', 'US East 1 zone'),
  ('us-west-2', 'US West 2 zone'),
  ('eu-west-1', 'Europe West 1 zone');

-- Insert default environments if they don't exist
INSERT OR IGNORE INTO environments (name) VALUES 
  ('production'),
  ('staging'),
  ('development'),
  ('testing');