import { Config } from '../../types'
import { Toggle } from '../ui/Toggle'

interface Props {
  config: Config
  onChange: (updates: Partial<Config>) => void
}

export function Step4Features({ config, onChange }: Props) {
  return (
    <div className="space-y-6">
      <Toggle
        checked={config.enable_queue_worker}
        onChange={(v) => onChange({ enable_queue_worker: v })}
        label="Enable Queue Worker"
        description="Sets up a Supervisor worker to process Laravel queue jobs"
      />
      <Toggle
        checked={config.enable_scheduler}
        onChange={(v) => onChange({ enable_scheduler: v })}
        label="Enable Task Scheduler"
        description="Adds a cron job to run Laravel's scheduled tasks every minute"
      />
    </div>
  )
}
