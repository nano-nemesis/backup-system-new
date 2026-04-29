export type Role = 'admin' | 'viewer'

export interface User {
  id: string
  username: string
  role: Role
}

export interface BackupFile {
  name: string
  rel_path: string
  size: number
  size_str: string
  mod_time: string // ISO 8601
}

export type NodeStatus = 'SUCCESS' | 'ERROR' | 'UNKNOWN'

export interface NodeSummary {
  name: string
  type: 'mikrotik' | 'database'
  host: string
  last_status: NodeStatus
  last_time: string | null // ISO 8601 or null
  last_msg: string
  backup_files: BackupFile[]
}

export interface NodesResponse {
  nodes: NodeSummary[]
  stats: {
    total: number
    ok: number
    failed: number
    unknown: number
  }
}

export interface NodeDetailResponse {
  node: NodeSummary
  logs: string[]
}

export interface AdminUser {
  id: string
  username: string
  role: Role
  created_at: string // ISO 8601
}

export interface AdminUsersResponse {
  users: AdminUser[]
}

export interface SetupStatusResponse {
  needs_setup: boolean
}

export interface MeResponse {
  user: User
}

export interface LoginResponse {
  user: User
}

export interface ParsedLogEntry {
  timestamp: string
  date: string
  status: 'SUCCESS' | 'ERROR' | null
  raw: string
}
