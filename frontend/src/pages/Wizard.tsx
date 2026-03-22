import { useCallback, useEffect, useState } from 'react'
import { Config } from '../types'
import { getConfig } from '../api/client'
import { Step1Domain } from '../components/wizard/Step1Domain'
import { Step2Database } from '../components/wizard/Step2Database'
import { Step3Server } from '../components/wizard/Step3Server'
import { Step4Features } from '../components/wizard/Step4Features'
import { Step5Review } from '../components/wizard/Step5Review'
import { Button } from '../components/ui/Button'

const STEPS = [
  { title: 'Domain & Project', subtitle: 'Configure your domain and repository' },
  { title: 'Database', subtitle: 'Set up MySQL credentials' },
  { title: 'PHP & Server', subtitle: 'Configure PHP and server user' },
  { title: 'Features', subtitle: 'Enable optional services' },
  { title: 'Review & Deploy', subtitle: 'Review settings and deploy' },
]

const emptyConfig: Config = {
  domain: '', github_repo: '', github_branch: 'main',
  php_version: '8.3', db_password: '', db_name: '', db_user: '', db_type: 'mysql',
  site_user: '', site_group: 'www-data',
  enable_queue_worker: true, enable_scheduler: true, dns_confirmed: false,
}

export function Wizard() {
  const [step, setStep] = useState(0)
  const [config, setConfig] = useState<Config>(emptyConfig)
  const [loading, setLoading] = useState(true)
  const [deployRunning, setDeployRunning] = useState(false)

  useEffect(() => {
    getConfig().then((cfg) => { setConfig(cfg); setLoading(false) }).catch(() => setLoading(false))
  }, [])

  const update = useCallback((updates: Partial<Config>) => setConfig((c) => ({ ...c, ...updates })), [])

  const canProceed = (): boolean => {
    switch (step) {
      case 0: return !!(config.domain && config.github_repo && config.github_branch)
      case 1: return !!(config.db_password && config.db_password.length >= 8 && config.db_name && config.db_user)
      case 2: return !!(config.php_version && config.site_user && config.site_group)
      case 3: return true
      default: return true
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-32">
        <div className="text-gray-500 text-sm">Loading configuration...</div>
      </div>
    )
  }

  const stepComponents = [
    <Step1Domain key="step1" config={config} onChange={update} />,
    <Step2Database key="step2" config={config} onChange={update} />,
    <Step3Server key="step3" config={config} onChange={update} />,
    <Step4Features key="step4" config={config} onChange={update} />,
    <Step5Review key="step5" config={config} onChange={update} deployRunning={deployRunning} setDeployRunning={setDeployRunning} />,
  ]

  return (
    <div className="max-w-xl mx-auto px-4 py-8">
      {/* Card container */}
      <div className="bg-gray-900/60 ring-1 ring-white/10 rounded-xl overflow-hidden">
        {/* Step Indicators */}
        <div className="px-6 pt-6 pb-4">
          <div className="flex items-center gap-1.5">
            {STEPS.map((_s, i) => (
              <div key={i} className="flex items-center gap-1.5 flex-1">
                <button
                  onClick={() => !deployRunning && setStep(i)}
                  disabled={deployRunning}
                  className={`flex items-center justify-center w-7 h-7 rounded-full text-[11px] font-bold transition-all
                    disabled:opacity-50 disabled:cursor-not-allowed
                    ${i === step
                      ? 'bg-indigo-600 text-white ring-2 ring-indigo-500/30 ring-offset-2 ring-offset-gray-900'
                      : i < step
                        ? 'bg-indigo-600/20 text-indigo-400'
                        : 'bg-white/5 text-gray-600'
                    }`}
                >
                  {i < step ? (
                    <svg viewBox="0 0 20 20" fill="currentColor" className="size-3.5">
                      <path fillRule="evenodd" clipRule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" />
                    </svg>
                  ) : (
                    i + 1
                  )}
                </button>
                {i < STEPS.length - 1 && (
                  <div className={`h-px flex-1 transition-colors ${i < step ? 'bg-indigo-600/40' : 'bg-white/5'}`} />
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Divider */}
        <div className="border-t border-white/5" />

        {/* Step Content */}
        <div className="px-6 py-6">
          <div className="mb-6">
            <h2 className="text-base font-semibold text-white">{STEPS[step].title}</h2>
            <p className="text-xs text-gray-500 mt-0.5">{STEPS[step].subtitle}</p>
          </div>
          {stepComponents[step]}
          {step < 4 && !canProceed() && (
            <p className="mt-4 text-[11px] text-yellow-400/70">Fill in all required fields to continue.</p>
          )}
        </div>

        {/* Navigation Footer */}
        {step < 4 && (
          <>
            <div className="border-t border-white/5" />
            <div className="px-6 py-4 flex justify-between">
              <Button variant="ghost" size="sm" onClick={() => setStep((s) => Math.max(0, s - 1))} disabled={step === 0}>
                Back
              </Button>
              <Button size="sm" onClick={() => setStep((s) => Math.min(4, s + 1))} disabled={!canProceed()}>
                Continue
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
