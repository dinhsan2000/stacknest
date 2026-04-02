export type ServiceName = 'apache' | 'nginx' | 'mysql' | 'postgres' | 'mongodb' | 'php' | 'redis'

export type ServiceStatus = 'running' | 'stopped' | 'starting' | 'stopping' | 'error'

export interface ServiceInfo {
  name: ServiceName
  display: string
  status: ServiceStatus
  port: number
  version: string
  pid: number
  error?: string
  enabled: boolean
  auto_recover: boolean
  uptime_since: number
  restart_count: number
  crash_loop: boolean
}

export interface VirtualHost {
  name: string
  domain: string
  root: string
  ssl: boolean
  active: boolean
  server: string
}

export interface ServiceConfig {
  enabled: boolean
  port: number
  path: string
  version: string
  auto_recover: boolean
}

export interface AppConfig {
  root_path: string
  bin_path: string
  data_path: string
  www_path: string
  log_path: string
  apache: ServiceConfig
  nginx: ServiceConfig
  mysql: ServiceConfig
  postgres: ServiceConfig
  mongodb: ServiceConfig
  php: ServiceConfig
  redis: ServiceConfig
  auto_start: boolean
  theme: 'light' | 'dark'
}

export interface VersionStatus {
  version: string
  installed: boolean
  active: boolean
  exe_path: string
}

export interface ServiceVersionStatus {
  service: string
  versions: VersionStatus[]
}

export interface VersionSpec {
  version: string
  url: string
  zip_strip: string
  exe_sub_dir: string
}

export interface ServiceCatalog {
  exe_name: string
  versions: VersionSpec[]
}

export interface BackupInfo {
  name: string
  size: number
  database: string
  created_at: string
}

export interface Project {
  id: string
  name: string
  doc_root: string
  domain: string
  server: string
  ssl: boolean
  php_path: string
  services: Record<string, boolean>
  created_at: string
  active: boolean
}
