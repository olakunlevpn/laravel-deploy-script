import { useState } from 'react'
import { Config, DeployStepStatus, PreflightResult } from '../../types'
import { Badge } from '../ui/Badge'
import { Button } from '../ui/Button'
import { CopyButton } from '../ui/CopyButton'
import { createDeployStream, saveConfig } from '../../api/client'

interface Props {
  config: Config
  onChange: (updates: Partial<Config>) => void
  deployRunning: boolean
  setDeployRunning: (v: boolean) => void
}

const STEP_NAMES = [
  '', 'Create Directory', 'Clone Repository', 'Setup Database',
  'Configure .env', 'Install Dependencies', 'Set Permissions',
  'Configure Nginx', 'Install SSL', 'Setup Queue Worker', 'Setup Scheduler',
  'Health Check',
]

const COPY_FIELDS = new Set(['Site Root', 'Database Name', 'Database User'])

function statusBadge(status: string) {
  if (status === 'success') return <Badge variant="success">Success</Badge>
  if (status === 'failed') return <Badge variant="error">Failed</Badge>
  if (status === 'running') return <Badge variant="info">Running...</Badge>
  if (status === 'skipped') return <Badge variant="warning">Skipped</Badge>
  return <Badge variant="default">Pending</Badge>
}

export function Step5Review({ config, onChange, deployRunning, setDeployRunning }: Props) {
  const [steps, setSteps] = useState<DeployStepStatus[]>([])
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')
  const [preflight, setPreflight] = useState<PreflightResult | null>(null)

  const siteRoot = `/home/${config.site_user}/${config.domain}`

  const summaryRows = [
    ['Domain', config.domain],
    ['GitHub Repo', config.github_repo],
    ['Branch', config.github_branch],
    ['PHP Version', config.php_version],
    ['Database Type', config.db_type || 'mysql'],
    ['Site User', config.site_user],
    ['Site Group', config.site_group],
    ['Site Root', siteRoot],
    ['Database Name', config.db_name],
    ['Database User', config.db_user],
    ['Queue Worker', config.enable_queue_worker ? 'Enabled' : 'Disabled'],
    ['Scheduler', config.enable_scheduler ? 'Enabled' : 'Disabled'],
  ]

  async function handleDeploy() {
    if (!config.domain || !config.github_repo || !config.db_password) {
      setError('Domain, GitHub repo, and database password are required.')
      return
    }
    setError('')
    setDeployRunning(true)
    setDone(false)
    setSteps([])
    setPreflight(null)

    try {
      await saveConfig(config)
    } catch {
      setError('Failed to save config. Check validation errors.')
      setDeployRunning(false)
      return
    }

    createDeployStream({
      onStep: (stepData: DeployStepStatus) => {
        setSteps((prev) => {
          const existing = prev.findIndex((s) => s.step === stepData.step)
          if (existing >= 0) {
            const updated = [...prev]
            updated[existing] = stepData
            return updated
          }
          return [...prev, stepData]
        })
      },
      onDone: () => {
        setDeployRunning(false)
        setDone(true)
      },
      onError: (message: string) => {
        setDeployRunning(false)
        setError(message)
      },
      onPreflight: (result: PreflightResult) => {
        setPreflight(result)
      },
    })
  }

  return (
    <div className="space-y-6">
      {/* Config Summary */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <tbody>
            {summaryRows.map(([label, value]) => (
              <tr key={label} className="border-b border-gray-700 last:border-0">
                <td className="px-4 py-2 text-gray-400 font-medium w-1/3">{label}</td>
                <td className="px-4 py-2 text-gray-100 font-mono">
                  {value || '—'}
                  {COPY_FIELDS.has(label) && value && <CopyButton text={value} />}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* DNS Confirmation */}
      <label className="flex items-start gap-3 cursor-pointer p-3 bg-yellow-900/20 border border-yellow-800 rounded-lg">
        <input
          type="checkbox"
          checked={config.dns_confirmed}
          onChange={(e) => onChange({ dns_confirmed: e.target.checked })}
          className="mt-0.5 h-4 w-4 rounded border-gray-600 bg-gray-700 text-indigo-600"
        />
        <span className="text-sm text-yellow-200">
          I have pointed DNS for <strong>{config.domain || 'your domain'}</strong> and{' '}
          <strong>www.{config.domain || 'your domain'}</strong> to this server's IP address.
          <span className="block text-yellow-400/70 text-xs mt-1">
            If unchecked, the SSL certificate step will be skipped.
          </span>
        </span>
      </label>

      {error && (
        <div className="p-3 bg-red-900/30 border border-red-800 rounded-lg text-red-300 text-sm">{error}</div>
      )}

      {/* Preflight Results */}
      {preflight && (
        <div className={`p-4 rounded-lg border ${preflight.passed ? 'bg-green-900/20 border-green-800' : 'bg-red-900/20 border-red-800'}`}>
          <h4 className="text-sm font-semibold mb-2 text-gray-300">
            {preflight.passed ? 'Preflight checks passed' : 'Preflight checks failed'}
          </h4>
          <div className="space-y-1">
            {preflight.checks.map((check, i) => (
              <div key={i} className="flex items-center gap-2 text-xs">
                <span className={check.passed ? 'text-green-400' : 'text-red-400'}>
                  {check.passed ? '\u2713' : '\u2717'}
                </span>
                <span className="text-gray-300">{check.name}</span>
                <span className="text-gray-500">{check.message}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Deploy Button */}
      {!deployRunning && !done && (
        <Button onClick={handleDeploy} className="w-full justify-center py-3 text-base">
          Deploy Now
        </Button>
      )}

      {/* Live Progress */}
      {(deployRunning || done) && steps.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Deployment Progress</h4>
          {Array.from({ length: 11 }, (_, i) => i + 1).map((n) => {
            const step = steps.find((s) => s.step === n)
            return (
              <div key={n} className="flex items-start gap-3 p-3 bg-gray-800 rounded-md">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-500">Step {n}</span>
                    <span className="text-sm text-gray-200">{step?.name || STEP_NAMES[n]}</span>
                    {step && statusBadge(step.status)}
                  </div>
                  {step?.output && step.status === 'failed' && (
                    <pre className="mt-2 text-xs text-red-300 bg-red-900/20 p-2 rounded overflow-auto max-h-32">
                      {step.output}
                    </pre>
                  )}
                </div>
              </div>
            )
          })}
          {done && (
            <div className="p-3 bg-green-900/30 border-green-800 rounded-lg text-green-300 text-sm">
              Deployment complete!{' '}
              <a href="/dashboard" className="underline font-medium">Go to Dashboard →</a>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
