import { Config } from '../../types'

interface Props {
  config: Config
  onChange: (updates: Partial<Config>) => void
}

export function Step1Domain({ config, onChange }: Props) {
  return (
    <div className="space-y-5">
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">Domain Name</label>
        <input
          type="text"
          placeholder="myapp.com"
          value={config.domain}
          onChange={(e) => onChange({ domain: e.target.value })}
          className="w-full bg-gray-800 border border-gray-700 rounded-md px-3 py-2 text-gray-100
            placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">GitHub Repository URL</label>
        <input
          type="text"
          placeholder="https://github.com/username/repo"
          value={config.github_repo}
          onChange={(e) => onChange({ github_repo: e.target.value })}
          className="w-full bg-gray-800 border border-gray-700 rounded-md px-3 py-2 text-gray-100
            placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-1">Branch</label>
        <input
          type="text"
          placeholder="main"
          value={config.github_branch}
          onChange={(e) => onChange({ github_branch: e.target.value })}
          className="w-full bg-gray-800 border border-gray-700 rounded-md px-3 py-2 text-gray-100
            placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        />
      </div>
    </div>
  )
}
