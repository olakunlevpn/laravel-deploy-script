import { useEffect, useState } from 'react'
import { Config } from '../../types'

function deriveDBName(domain: string): string {
  return domain
    .replace(/[^a-zA-Z0-9]+/g, '_')
    .replace(/_com$/, '')
    .replace(/_+/g, '_')
    .replace(/^_|_$/g, '')
}

interface Props {
  config: Config
  onChange: (updates: Partial<Config>) => void
}

export function Step2Database({ config, onChange }: Props) {
  const [showPassword, setShowPassword] = useState(false)
  const [dbNameManual, setDbNameManual] = useState(false)
  const [dbUserManual, setDbUserManual] = useState(false)

  // Reactively derive DB name from domain unless manually edited
  useEffect(() => {
    if (!dbNameManual && config.domain) {
      const derived = deriveDBName(config.domain)
      onChange({ db_name: derived, ...(!dbUserManual ? { db_user: derived + '_user' } : {}) })
    }
  }, [config.domain])

  // Reactively derive DB user from DB name unless manually edited
  useEffect(() => {
    if (!dbUserManual && config.db_name) {
      onChange({ db_user: config.db_name + '_user' })
    }
  }, [config.db_name])

  const inputClass = "w-full bg-gray-800 border border-gray-700 rounded-md px-3 py-2 text-gray-100 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"

  return (
    <div className="space-y-5">
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">Database Type</label>
        <select
          value={config.db_type || 'mysql'}
          onChange={(e) => onChange({ db_type: e.target.value })}
          className={inputClass}
        >
          <option value="mysql">MySQL / MariaDB</option>
          <option value="postgresql">PostgreSQL</option>
        </select>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">Database Password</label>
        <div className="relative">
          <input
            type={showPassword ? 'text' : 'password'}
            placeholder="Strong password"
            value={config.db_password}
            onChange={(e) => onChange({ db_password: e.target.value })}
            className={inputClass + ' pr-20'}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400 hover:text-gray-200"
          >
            {showPassword ? 'Hide' : 'Show'}
          </button>
        </div>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">
          Database Name
          <span className="ml-2 text-xs text-gray-500">(auto-generated from domain)</span>
        </label>
        <input
          type="text"
          value={config.db_name}
          onChange={(e) => { setDbNameManual(true); onChange({ db_name: e.target.value }) }}
          className={inputClass}
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">
          Database User
          <span className="ml-2 text-xs text-gray-500">(auto-generated)</span>
        </label>
        <input
          type="text"
          value={config.db_user}
          onChange={(e) => { setDbUserManual(true); onChange({ db_user: e.target.value }) }}
          className={inputClass}
        />
      </div>
    </div>
  )
}
