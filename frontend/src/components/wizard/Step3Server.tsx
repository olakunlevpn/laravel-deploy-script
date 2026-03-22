import { Config } from '../../types'

interface Props {
  config: Config
  onChange: (updates: Partial<Config>) => void
}

export function Step3Server({ config, onChange }: Props) {
  const inputClass = "w-full bg-gray-800 border border-gray-700 rounded-md px-3 py-2 text-gray-100 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"

  return (
    <div className="space-y-5">
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">PHP Version</label>
        <select
          value={config.php_version}
          onChange={(e) => onChange({ php_version: e.target.value })}
          className={inputClass}
        >
          <option value="8.1">PHP 8.1</option>
          <option value="8.2">PHP 8.2</option>
          <option value="8.3">PHP 8.3</option>
        </select>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">
          Site User
          <span className="ml-2 text-xs text-gray-500">(auto-detected from server)</span>
        </label>
        <input
          type="text"
          value={config.site_user}
          onChange={(e) => onChange({ site_user: e.target.value })}
          className={inputClass}
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">Site Group</label>
        <input
          type="text"
          value={config.site_group}
          onChange={(e) => onChange({ site_group: e.target.value })}
          className={inputClass}
        />
      </div>
      {config.site_user && config.domain && (
        <div className="bg-gray-800 border border-gray-700 rounded-md px-3 py-2">
          <span className="text-xs text-gray-500">Site root will be: </span>
          <span className="text-xs text-indigo-400 font-mono">
            /home/{config.site_user}/{config.domain}
          </span>
        </div>
      )}
    </div>
  )
}
