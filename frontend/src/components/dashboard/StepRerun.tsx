import { useEffect, useState } from 'react'
import { rerunStep, getDeployStatus } from '../../api/client'
import { DeployStepStatus } from '../../types'
import { Card } from '../ui/Card'
import { Badge } from '../ui/Badge'
import { Button } from '../ui/Button'

const STEP_NAMES = [
  '', 'Create Directory', 'Clone Repository', 'Setup Database',
  'Configure .env', 'Install Dependencies', 'Set Permissions',
  'Configure Nginx', 'Install SSL', 'Setup Queue Worker', 'Setup Scheduler',
  'Health Check',
]

function statusBadge(status?: string) {
  if (!status) return <Badge variant="default">Never run</Badge>
  if (status === 'success') return <Badge variant="success">Success</Badge>
  if (status === 'failed') return <Badge variant="error">Failed</Badge>
  if (status === 'skipped') return <Badge variant="warning">Skipped</Badge>
  return <Badge variant="default">{status}</Badge>
}

export function StepRerun() {
  const [steps, setSteps] = useState<Record<number, DeployStepStatus>>({})
  const [running, setRunning] = useState<number | null>(null)
  const [output, setOutput] = useState<{ step: number; text: string } | null>(null)

  useEffect(() => {
    getDeployStatus().then((s) => {
      const map: Record<number, DeployStepStatus> = {}
      s.steps.forEach((st) => { map[st.step] = st })
      setSteps(map)
    }).catch(() => {})
  }, [])

  async function handleRerun(n: number) {
    setRunning(n)
    setOutput(null)
    try {
      const result = await rerunStep(n)
      setOutput({ step: n, text: result.output })
      setSteps((prev) => ({
        ...prev,
        [n]: { step: n, name: STEP_NAMES[n], status: result.success ? 'success' : 'failed', output: result.output, timestamp: new Date().toISOString() },
      }))
    } catch {
      setOutput({ step: n, text: 'Request failed. Check server connectivity.' })
    } finally {
      setRunning(null)
    }
  }

  return (
    <Card title="Deployment Steps">
      <div className="space-y-2">
        {Array.from({ length: 11 }, (_, i) => i + 1).map((n) => (
          <div key={n} className="flex items-center gap-3 p-3 bg-gray-800 rounded-md">
            <div className="w-6 h-6 flex items-center justify-center rounded-full bg-gray-700 text-xs text-gray-400 flex-shrink-0">
              {n}
            </div>
            <div className="flex-1">
              <div className="text-sm text-gray-200">{STEP_NAMES[n]}</div>
              <div className="mt-1 flex items-center gap-2">
                {statusBadge(steps[n]?.status)}
                {steps[n]?.timestamp && (
                  <span className="text-xs text-gray-500">
                    {new Date(steps[n].timestamp).toLocaleString()}
                  </span>
                )}
              </div>
            </div>
            <Button
              variant="ghost"
              loading={running === n}
              onClick={() => handleRerun(n)}
              className="text-xs"
            >
              Re-run
            </Button>
          </div>
        ))}
      </div>
      {output && (
        <div className="mt-4">
          <div className="text-xs text-gray-500 mb-1">Output from Step {output.step}:</div>
          <pre className="bg-gray-950 text-xs text-gray-300 p-3 rounded-md overflow-auto max-h-48">{output.text}</pre>
        </div>
      )}
    </Card>
  )
}
