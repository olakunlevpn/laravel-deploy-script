import { useEffect, useState, ReactNode } from 'react'
import { StatusResponse } from '../../types'
import { getStatus, nginxAction, supervisorAction, queueWorkerAction, sslRenew } from '../../api/client'
import { Button } from '../ui/Button'

interface ServiceCardProps {
  name: string
  running: boolean
  actions?: ReactNode
}

function ServiceCard({ name, running, actions }: ServiceCardProps) {
  return (
    <div className="bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg px-4 py-3.5">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          <div className={`flex-none rounded-full p-1 ${running ? 'bg-green-500/10' : 'bg-red-500/10'}`}>
            <div className={`size-2 rounded-full ${running ? 'bg-green-500' : 'bg-red-500'}`} />
          </div>
          <div className="min-w-0">
            <div className="text-sm font-medium text-gray-200 truncate">{name}</div>
            <div className="text-[11px] text-gray-500 mt-0.5">{running ? 'Running' : 'Stopped'}</div>
          </div>
        </div>
        {actions && <div className="flex gap-1.5 shrink-0">{actions}</div>}
      </div>
    </div>
  )
}

export function StatusCards() {
  const [status, setStatus] = useState<StatusResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [busyKey, setBusyKey] = useState('')
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const refresh = () => {
    setLoading(true)
    getStatus().then((s) => { setStatus(s); setLastRefresh(new Date()) }).finally(() => setLoading(false))
  }

  useEffect(() => { refresh() }, [])

  useEffect(() => {
    const interval = setInterval(refresh, 30000)
    return () => clearInterval(interval)
  }, [])

  async function act(key: string, fn: () => Promise<unknown>) {
    setBusyKey(key)
    try {
      await fn()
    } catch {
      // refresh will show real state
    } finally {
      setBusyKey('')
      setTimeout(refresh, 1500)
    }
  }

  if (loading && !status) {
    return <div className="text-gray-500 text-sm py-8">Loading service status...</div>
  }

  if (!status) return null

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-medium text-gray-300">Services</h2>
        <div className="flex items-center gap-3">
          <span className="text-[11px] text-gray-600">{lastRefresh.toLocaleTimeString()}</span>
          <Button variant="ghost" size="sm" onClick={refresh}>Refresh</Button>
        </div>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-3">
        <ServiceCard
          name="Nginx"
          running={status.nginx.running}
          actions={
            <Button size="sm" variant="ghost" loading={busyKey === 'nginx-restart'}
              onClick={() => act('nginx-restart', () => nginxAction('restart'))}>
              Restart
            </Button>
          }
        />
        <ServiceCard name="PHP-FPM" running={status.php_fpm.running} />
        <ServiceCard name="MySQL" running={status.mysql.running} />
        <ServiceCard
          name="Supervisor"
          running={status.supervisor.running}
          actions={
            <div className="flex gap-1">
              <Button size="sm" variant="ghost" loading={busyKey === 'sup-restart'}
                onClick={() => act('sup-restart', () => supervisorAction('restart'))}>Restart</Button>
            </div>
          }
        />
        <div className="bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg px-4 py-3.5">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className={`flex-none rounded-full p-1 ${status.ssl.valid ? 'bg-green-500/10' : 'bg-red-500/10'}`}>
                <div className={`size-2 rounded-full ${status.ssl.valid ? 'bg-green-500' : 'bg-red-500'}`} />
              </div>
              <div>
                <div className="text-sm font-medium text-gray-200">SSL Certificate</div>
                <div className="text-[11px] mt-0.5">
                  {status.ssl.valid ? (
                    <span className={status.ssl.days_left > 30 ? 'text-gray-500' : 'text-yellow-400'}>
                      Expires {status.ssl.expiry_date} ({status.ssl.days_left}d)
                    </span>
                  ) : (
                    <span className="text-red-400">Not valid</span>
                  )}
                </div>
              </div>
            </div>
            <Button size="sm" variant="ghost" loading={busyKey === 'ssl-renew'}
              onClick={() => act('ssl-renew', sslRenew)}>
              Renew
            </Button>
          </div>
        </div>
        <ServiceCard
          name="Queue Worker"
          running={status.queue_worker.running}
          actions={
            <div className="flex gap-1">
              <Button size="sm" variant="ghost" loading={busyKey === 'qw-restart'}
                onClick={() => act('qw-restart', () => queueWorkerAction('restart'))}>Restart</Button>
            </div>
          }
        />
      </div>
    </div>
  )
}
