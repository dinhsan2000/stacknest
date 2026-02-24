export type ServiceName = 'apache' | 'nginx' | 'mysql' | 'php' | 'redis'

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
}

export interface VirtualHost {
  name: string
  domain: string
  root: string
  ssl: boolean
  active: boolean
}

export interface ServiceConfig {
  enabled: boolean
  port: number
  path: string
  version: string
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
