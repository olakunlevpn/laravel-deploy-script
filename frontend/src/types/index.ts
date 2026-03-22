export interface Config {
  domain: string
  github_repo: string
  github_branch: string
  php_version: string
  db_password: string
  db_name: string
  db_user: string
  db_type: string
  site_user: string
  site_group: string
  enable_queue_worker: boolean
  enable_scheduler: boolean
  dns_confirmed: boolean
}

export interface PreflightCheck {
  name: string
  passed: boolean
  message: string
}

export interface PreflightResult {
  passed: boolean
  checks: PreflightCheck[]
}

export interface ServiceStatus {
  name: string
  running: boolean
}

export interface SSLStatus {
  expiry_date: string
  days_left: number
  valid: boolean
}

export interface StatusResponse {
  nginx: ServiceStatus
  php_fpm: ServiceStatus
  mysql: ServiceStatus
  supervisor: ServiceStatus
  ssl: SSLStatus
  queue_worker: ServiceStatus
  server_ip: string
}

export interface DeployStepStatus {
  step: number
  name: string
  status: 'pending' | 'running' | 'success' | 'failed' | 'skipped'
  output: string
  timestamp: string
}

export interface DeployStatus {
  running: boolean
  steps: DeployStepStatus[]
}

export interface ActionResult {
  success: boolean
  output: string
}
