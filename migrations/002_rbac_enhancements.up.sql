-- RBAC Enhancements Migration
-- This migration adds indexes and constraints for better RBAC performance

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_namespace_id ON user_roles(namespace_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

-- Add indexes for group-based RBAC
CREATE INDEX IF NOT EXISTS idx_user_groups_user_id ON user_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_user_groups_group_id ON user_groups(group_id);
CREATE INDEX IF NOT EXISTS idx_group_roles_group_id ON group_roles(group_id);
CREATE INDEX IF NOT EXISTS idx_group_roles_role_id ON group_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_group_roles_namespace_id ON group_roles(namespace_id);

-- Add audit trail for RBAC changes
CREATE TABLE rbac_audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  action TEXT NOT NULL, -- 'ASSIGN_ROLE', 'REMOVE_ROLE', 'CREATE_ROLE', 'DELETE_ROLE'
  actor_user_id INTEGER REFERENCES users(id),
  target_user_id INTEGER REFERENCES users(id),
  role_id INTEGER REFERENCES roles(id),
  namespace_id INTEGER REFERENCES namespaces(id),
  details TEXT, -- JSON with additional details
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rbac_audit_log_action ON rbac_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_log_actor ON rbac_audit_log(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_log_target ON rbac_audit_log(target_user_id);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_log_created_at ON rbac_audit_log(created_at);

-- Add role hierarchy support (optional for future use)
CREATE TABLE role_hierarchy (
  parent_role_id INTEGER NOT NULL REFERENCES roles(id),
  child_role_id INTEGER NOT NULL REFERENCES roles(id),
  PRIMARY KEY (parent_role_id, child_role_id)
);

-- Add permissions table for fine-grained access control
CREATE TABLE permissions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT,
  resource TEXT NOT NULL, -- e.g., 'secrets', 'users', 'roles'
  action TEXT NOT NULL,   -- e.g., 'read', 'write', 'delete', 'admin'
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE role_permissions (
  role_id INTEGER NOT NULL REFERENCES roles(id),
  permission_id INTEGER NOT NULL REFERENCES permissions(id),
  PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);