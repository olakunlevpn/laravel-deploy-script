import { useEffect, useState } from 'react'
import { Config } from '../../types'
import { getConfig } from '../../api/client'
import { Card } from '../ui/Card'
import { CopyButton } from '../ui/CopyButton'

export function SystemdHelper() {
  const [config, setConfig] = useState<Config | null>(null)

  useEffect(() => {
    getConfig().then(setConfig).catch(() => {})
  }, [])

  if (!config) return null

  const serviceFile = `[Unit]
Description=Laravel Deploy Panel
After=network.target

[Service]
Type=simple
ExecStart=/opt/deploy-panel/panel --port 4432
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target`

  const instructions = `sudo tee /etc/systemd/system/deploy-panel.service << 'EOF'
${serviceFile}
EOF
sudo systemctl daemon-reload
sudo systemctl enable deploy-panel
sudo systemctl start deploy-panel`

  return (
    <Card title="Auto-start on Boot">
      <p className="text-xs text-gray-500 mb-3">
        Run these commands to install the panel as a systemd service:
      </p>
      <div className="relative">
        <pre className="bg-gray-950 text-gray-300 text-xs p-3 rounded-md overflow-auto max-h-48 font-mono">
          {instructions}
        </pre>
        <div className="absolute top-2 right-2">
          <CopyButton text={instructions} />
        </div>
      </div>
    </Card>
  )
}
