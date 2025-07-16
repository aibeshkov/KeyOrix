-- 🌐 Справочники: неймспейсы, зоны, окружения

CREATE TABLE namespaces (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE zones (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE environments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

-- 🔐 Секреты и версии

CREATE TABLE secret_nodes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  parent_id INTEGER REFERENCES secret_nodes(id) ON DELETE CASCADE,
  namespace_id INTEGER NOT NULL REFERENCES namespaces(id),
  zone_id INTEGER NOT NULL REFERENCES zones(id),
  environment_id INTEGER NOT NULL REFERENCES environments(id),
  name TEXT NOT NULL,
  is_secret BOOLEAN NOT NULL DEFAULT 0,
  type TEXT,
  max_reads INTEGER,
  expiration TIMESTAMP,
  metadata TEXT,
  status TEXT DEFAULT 'active',
  created_by TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE secret_versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  secret_node_id INTEGER NOT NULL REFERENCES secret_nodes(id) ON DELETE CASCADE,
  version_number INTEGER NOT NULL,
  encrypted_value BLOB NOT NULL,
  encryption_metadata TEXT,
  read_count INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE secret_access_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  secret_node_id INTEGER NOT NULL REFERENCES secret_nodes(id),
  secret_version_id INTEGER NOT NULL REFERENCES secret_versions(id),
  accessed_by TEXT,
  access_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  action TEXT,
  ip_address TEXT,
  user_agent TEXT
);

CREATE TABLE secret_metadata_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  secret_node_id INTEGER NOT NULL REFERENCES secret_nodes(id),
  changed_by TEXT,
  change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  old_metadata TEXT,
  new_metadata TEXT
);

-- 🧑‍💻 Пользователи, роли, группы

CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  email TEXT,
  password_hash TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE roles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

CREATE TABLE user_roles (
  user_id INTEGER NOT NULL REFERENCES users(id),
  role_id INTEGER NOT NULL REFERENCES roles(id),
  namespace_id INTEGER REFERENCES namespaces(id),
  PRIMARY KEY (user_id, role_id, namespace_id)
);

CREATE TABLE groups (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

CREATE TABLE user_groups (
  user_id INTEGER NOT NULL REFERENCES users(id),
  group_id INTEGER NOT NULL REFERENCES groups(id),
  PRIMARY KEY (user_id, group_id)
);

CREATE TABLE group_roles (
  group_id INTEGER NOT NULL REFERENCES groups(id),
  role_id INTEGER NOT NULL REFERENCES roles(id),
  namespace_id INTEGER REFERENCES namespaces(id),
  PRIMARY KEY (group_id, role_id, namespace_id)
);

-- 🛡️ Аутентификация

CREATE TABLE sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES users(id),
  session_token TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP
);

CREATE TABLE password_resets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES users(id),
  token TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 🏷️ Теги

CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE secret_tags (
  secret_node_id INTEGER NOT NULL REFERENCES secret_nodes(id),
  tag_id INTEGER NOT NULL REFERENCES tags(id),
  PRIMARY KEY (secret_node_id, tag_id)
);

-- 📬 Уведомления

CREATE TABLE notifications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES users(id),
  secret_node_id INTEGER REFERENCES secret_nodes(id),
  type TEXT NOT NULL,
  message TEXT NOT NULL,
  is_read BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  user_id INTEGER REFERENCES users(id),
  secret_node_id INTEGER REFERENCES secret_nodes(id),
  description TEXT,
  event_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ⚙️ Настройки

CREATE TABLE settings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER REFERENCES users(id),
  key TEXT NOT NULL,
  value TEXT,
  UNIQUE(user_id, key)
);

CREATE TABLE system_metadata (
  key TEXT PRIMARY KEY,
  value TEXT,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 🔐 API и интеграции

CREATE TABLE api_clients (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  client_id TEXT NOT NULL UNIQUE,
  client_secret TEXT NOT NULL,
  scopes TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE api_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  client_id INTEGER NOT NULL REFERENCES api_clients(id),
  user_id INTEGER REFERENCES users(id),
  token TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  scope TEXT,
  revoked BOOLEAN DEFAULT FALSE
);

CREATE TABLE rate_limits (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  client_id INTEGER NOT NULL REFERENCES api_clients(id),
  method TEXT NOT NULL,
  limit_per_minute INTEGER NOT NULL DEFAULT 60,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE api_call_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  client_id INTEGER REFERENCES api_clients(id),
  user_id INTEGER REFERENCES users(id),
  method TEXT,
  path TEXT,
  status_code INTEGER,
  duration_ms INTEGER,
  ip_address TEXT,
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 🧠 gRPC

CREATE TABLE grpc_services (
  name TEXT PRIMARY KEY,
  version TEXT,
  description TEXT
);

-- 🌐 IdentityProvider (OIDC/LDAP/SSO)

CREATE TABLE identity_providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL,
  config TEXT NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE external_identities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider_id INTEGER NOT NULL REFERENCES identity_providers(id),
  user_id INTEGER NOT NULL REFERENCES users(id),
  external_id TEXT NOT NULL,
  email TEXT,
  name TEXT,
  metadata TEXT,
  linked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
