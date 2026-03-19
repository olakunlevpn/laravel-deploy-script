import { useState } from 'react'
import { laravelAction, nginxAction, resetPermissions } from '../../api/client'
import { ActionResult } from '../../types'
import { Card } from '../ui/Card'
import { Button } from '../ui/Button'

function useAction() {
  const [busy, setBusy] = useState('')
  const [result, setResult] = useState<ActionResult | null>(null)

  async function run(key: string, fn: () => Promise<ActionResult>) {
    setBusy(key)
    setResult(null)
    try {
      const res = await fn()
      setResult(res)
    } catch {
      setResult({ success: false, output: 'Request failed. Check server connectivity.' })
    } finally {
      setBusy('')
    }
  }

  return { busy, result, run }
}

export function ActionPanel() {
  const { busy, result, run } = useAction()

  const laravelActions = [
    { key: 'cache:clear', label: 'Clear Cache' },
    { key: 'config:clear', label: 'Clear Config' },
    { key: 'route:clear', label: 'Clear Routes' },
    { key: 'view:clear', label: 'Clear Views' },
    { key: 'migrate', label: 'Run Migrations' },
    { key: 'migrate:rollback', label: 'Rollback Migration' },
    { key: 'optimize', label: 'Optimize' },
    { key: 'storage:link', label: 'Storage Link' },
  ]

  return (
    <div className="space-y-4">
      {result && (
        <div className={`p-3 rounded-lg text-sm border ${result.success ? 'bg-green-900/20 border-green-800 text-green-300' : 'bg-red-900/20 border-red-800 text-red-300'}`}>
          <pre className="whitespace-pre-wrap text-xs">{result.output || (result.success ? 'Done' : 'Failed')}</pre>
        </div>
      )}

      <Card title="Laravel">
        <div className="flex flex-wrap gap-2">
          {laravelActions.map((a) => (
            <Button
              key={a.key}
              variant="secondary"
              loading={busy === a.key}
              onClick={() => run(a.key, () => laravelAction(a.key))}
            >
              {a.label}
            </Button>
          ))}
        </div>
      </Card>

      <Card title="Nginx">
        <div className="flex gap-2">
          <Button variant="secondary" loading={busy === 'nginx-reload'}
            onClick={() => run('nginx-reload', () => nginxAction('reload'))}>Reload</Button>
          <Button variant="secondary" loading={busy === 'nginx-restart'}
            onClick={() => run('nginx-restart', () => nginxAction('restart'))}>Restart</Button>
        </div>
      </Card>

      <Card title="Permissions">
        <Button variant="secondary" loading={busy === 'permissions'}
          onClick={() => run('permissions', resetPermissions)}>
          Re-apply File Permissions
        </Button>
      </Card>
    </div>
  )
}
