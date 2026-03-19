interface ToggleProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label: string
  description?: string
}

export function Toggle({ checked, onChange, label, description }: ToggleProps) {
  return (
    <label className="flex items-start gap-3 cursor-pointer select-none">
      <div className="relative mt-0.5 flex-shrink-0">
        <input
          type="checkbox"
          className="sr-only"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
        />
        <div className={`w-10 h-6 rounded-full transition-colors ${checked ? 'bg-indigo-600' : 'bg-gray-700'}`} />
        <div className={`absolute top-1 left-1 w-4 h-4 rounded-full bg-white transition-transform ${checked ? 'translate-x-4' : ''}`} />
      </div>
      <div>
        <div className="text-sm font-medium text-gray-200">{label}</div>
        {description && <div className="text-xs text-gray-500 mt-0.5">{description}</div>}
      </div>
    </label>
  )
}
