import type { Config, StatusResponse, DeployStatus, ActionResult, DeployStepStatus, PreflightResult } from '../types'

const BASE = ''  // Same origin

export async function getConfig(): Promise<Config> {
  const res = await fetch(`${BASE}/api/config`)
  if (!res.ok) throw new Error('Failed to load config')
  return res.json()
}

export async function saveConfig(cfg: Config): Promise<void> {
  const res = await fetch(`${BASE}/api/config`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(cfg),
  })
  if (!res.ok) throw new Error('Failed to save config')
}

export async function getStatus(): Promise<StatusResponse> {
  const res = await fetch(`${BASE}/api/status`)
  if (!res.ok) throw new Error('Failed to load status')
  return res.json()
}

export async function getDeployStatus(): Promise<DeployStatus> {
  const res = await fetch(`${BASE}/api/deploy/status`)
  if (!res.ok) throw new Error('Failed to load deploy status')
  return res.json()
}

export async function rerunStep(step: number): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/deploy/step/${step}`, { method: 'POST' })
  if (!res.ok) throw new Error(`Failed to rerun step ${step}: ${res.statusText}`)
  return res.json()
}

export async function nginxAction(action: 'reload' | 'restart'): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/actions/nginx/${action}`, { method: 'POST' })
  return res.json()
}

export async function supervisorAction(action: 'start' | 'stop' | 'restart'): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/actions/supervisor/${action}`, { method: 'POST' })
  return res.json()
}

export async function queueWorkerAction(action: 'start' | 'stop' | 'restart'): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/actions/queue-worker/${action}`, { method: 'POST' })
  return res.json()
}

export async function sslRenew(): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/actions/ssl/renew`, { method: 'POST' })
  return res.json()
}

export async function resetPermissions(): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/actions/permissions`, { method: 'POST' })
  return res.json()
}

export async function laravelAction(action: string): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/actions/laravel/${action}`, { method: 'POST' })
  return res.json()
}

export async function getLogs(): Promise<{ lines: string[]; error?: string }> {
  const res = await fetch(`${BASE}/api/logs/laravel`)
  return res.json()
}

export async function clearLogs(): Promise<ActionResult> {
  const res = await fetch(`${BASE}/api/logs/laravel`, { method: 'POST' })
  return res.json()
}

interface DeployStreamCallbacks {
  onStep: (data: DeployStepStatus) => void
  onDone: () => void
  onError?: (message: string) => void
  onPreflight?: (result: PreflightResult) => void
}

export function createDeployStream(callbacks: DeployStreamCallbacks): EventSource {
  const es = new EventSource(`${BASE}/api/deploy/stream`)
  es.onmessage = (e) => {
    const data = JSON.parse(e.data)
    if (data.done) {
      callbacks.onDone()
      es.close()
    } else if (data.error) {
      callbacks.onError?.(data.error)
      es.close()
    } else if (data.preflight) {
      callbacks.onPreflight?.(data.result as PreflightResult)
    } else if (data.step) {
      callbacks.onStep(data as DeployStepStatus)
    }
  }
  es.onerror = () => {
    es.close()
    callbacks.onError?.('Connection to server lost. Please check the server and try again.')
  }
  return es
}
