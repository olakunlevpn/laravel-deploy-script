import { useEffect, useState } from 'react'
import { StatusCards } from '../components/dashboard/StatusCards'
import { CopyButton } from '../components/ui/CopyButton'

interface ServerInfo {
  hostname: string
  os: string
  kernel: string
  uptime: string
  cpu_cores: number
  load_average: string
  memory_total: string
  memory_used: string
  memory_free: string
  memory_pct: string
  disk_total: string
  disk_used: string
  disk_free: string
  disk_pct: string
  php_version: string
  nginx_version: string
  db_version: string
  composer_version: string
  server_ip: string
  domain: string
  site_root: string
  site_user: string
  db_type: string
  db_name: string
  db_user: string
  github_repo: string
  github_branch: string
  php_config: string
  queue_worker: boolean
  scheduler: boolean
}

function InfoItem({ label, value, mono, copy }: { label: string; value: string; mono?: boolean; copy?: boolean }) {
  if (!value || value === '—') return (
    <div className="py-2.5 px-3">
      <div className="text-[11px] text-gray-500 mb-0.5">{label}</div>
      <div className="text-sm text-gray-600">—</div>
    </div>
  )
  return (
    <div className="py-2.5 px-3">
      <div className="text-[11px] text-gray-500 mb-0.5">{label}</div>
      <div className={`text-sm text-gray-200 ${mono ? 'font-mono' : ''} flex items-center gap-1.5`}>
        <span className="truncate">{value}</span>
        {copy && <CopyButton text={value} />}
      </div>
    </div>
  )
}

function UsageBar({ label, used, total, pct }: { label: string; used: string; total: string; pct: string }) {
  const numPct = parseInt(pct) || 0
  const barColor = numPct > 90 ? 'bg-red-500' : numPct > 70 ? 'bg-yellow-500' : 'bg-indigo-500'

  return (
    <div className="py-2.5 px-3">
      <div className="flex items-center justify-between mb-1.5">
        <span className="text-[11px] text-gray-500">{label}</span>
        <span className="text-[11px] text-gray-400">{used} / {total}</span>
      </div>
      <div className="h-1.5 bg-white/5 rounded-full overflow-hidden">
        <div className={`h-full rounded-full transition-all ${barColor}`} style={{ width: `${numPct}%` }} />
      </div>
      <div className="text-right text-[11px] text-gray-500 mt-1">{pct} used</div>
    </div>
  )
}

function VersionItem({ label, version }: { label: string; version: string }) {
  const installed = version && version !== 'Not installed'
  return (
    <div className="flex items-center gap-3 py-2 px-3">
      <div className={`flex-none rounded-full p-0.5 ${installed ? 'bg-green-500/10' : 'bg-red-500/10'}`}>
        <div className={`size-1.5 rounded-full ${installed ? 'bg-green-500' : 'bg-red-500'}`} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm text-gray-200">{label}</div>
        <div className="text-[11px] text-gray-500 truncate">{version || 'Not installed'}</div>
      </div>
    </div>
  )
}

export function Dashboard() {
  const [info, setInfo] = useState<ServerInfo | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/server-info')
      .then(res => res.json())
      .then(setInfo)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <div className="px-4 py-8 sm:px-6 text-gray-500 text-sm">Loading server information...</div>
  }

  if (!info) {
    return <div className="px-4 py-8 sm:px-6 text-gray-500 text-sm">Failed to load server information.</div>
  }

  return (
    <div className="px-4 py-6 sm:px-6 space-y-6">
      {/* Top row: System + Resources */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* System Info */}
        <div className="bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg overflow-hidden">
          <div className="px-3 pt-3 pb-2 border-b border-white/5">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">System</h3>
          </div>
          <div className="divide-y divide-white/5">
            <InfoItem label="Hostname" value={info.hostname} mono copy />
            <InfoItem label="Operating System" value={info.os} />
            <InfoItem label="Kernel" value={info.kernel} mono />
            <InfoItem label="Uptime" value={info.uptime} />
            <InfoItem label="CPU Cores" value={String(info.cpu_cores)} />
            <InfoItem label="Load Average" value={info.load_average} mono />
          </div>
        </div>

        {/* Resources */}
        <div className="bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg overflow-hidden">
          <div className="px-3 pt-3 pb-2 border-b border-white/5">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Resources</h3>
          </div>
          <div className="divide-y divide-white/5">
            <UsageBar label="Memory" used={info.memory_used} total={info.memory_total} pct={info.memory_pct} />
            <UsageBar label="Disk" used={info.disk_used} total={info.disk_total} pct={info.disk_pct} />
            <InfoItem label="Server IP" value={info.server_ip} mono copy />
          </div>
        </div>

        {/* Software */}
        <div className="bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg overflow-hidden">
          <div className="px-3 pt-3 pb-2 border-b border-white/5">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Software</h3>
          </div>
          <div className="divide-y divide-white/5">
            <VersionItem label="PHP" version={info.php_version} />
            <VersionItem label="Nginx" version={info.nginx_version} />
            <VersionItem label={info.db_type === 'postgresql' ? 'PostgreSQL' : 'MySQL'} version={info.db_version} />
            <VersionItem label="Composer" version={info.composer_version} />
          </div>
        </div>
      </div>

      {/* Project Info */}
      {info.domain && (
        <div className="bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg overflow-hidden">
          <div className="px-3 pt-3 pb-2 border-b border-white/5">
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Project</h3>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 divide-x divide-white/5">
            <InfoItem label="Domain" value={info.domain} mono copy />
            <InfoItem label="Site Root" value={info.site_root} mono copy />
            <InfoItem label="Site User" value={info.site_user} mono />
            <InfoItem label="Repository" value={info.github_repo} mono />
            <InfoItem label="Branch" value={info.github_branch} mono />
            <div className="py-2.5 px-3">
              <div className="text-[11px] text-gray-500 mb-1">Features</div>
              <div className="flex flex-wrap gap-1.5">
                <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${info.queue_worker ? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20' : 'bg-white/5 text-gray-500 ring-1 ring-white/10'}`}>
                  Queue
                </span>
                <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${info.scheduler ? 'bg-indigo-500/10 text-indigo-400 ring-1 ring-indigo-500/20' : 'bg-white/5 text-gray-500 ring-1 ring-white/10'}`}>
                  Cron
                </span>
              </div>
            </div>
          </div>
          <div className="border-t border-white/5 grid grid-cols-2 sm:grid-cols-3 divide-x divide-white/5">
            <InfoItem label="Database" value={`${info.db_name} (${info.db_type || 'mysql'})`} mono />
            <InfoItem label="DB User" value={info.db_user} mono copy />
            <InfoItem label="PHP Version" value={info.php_config} mono />
          </div>
        </div>
      )}

      {/* Service Status */}
      <StatusCards />
    </div>
  )
}
