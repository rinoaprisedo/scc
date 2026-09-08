import { useEffect, useRef, useState } from 'react'
import { CheckCircle2, X } from 'lucide-react'
import FormTab from './FormTab'
import ShirtSizeTab from './ShirtSizeTab'
import Button from '../ui/Button'
import { isFormComplete, isShirtComplete } from '../../utils/onboarding'
import { getKotaAsalOptions, getBandaraOptions, getBlazerSizeOptions } from '../../api/refData'

const FORM_ID = 'onboarding-active-form'

function TabHeader({ label, complete, active, disabled, onClick }) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={`flex flex-1 flex-col items-center gap-1 pb-3 text-center transition ${
        disabled ? 'cursor-not-allowed opacity-40' : ''
      }`}
    >
      <span className={`text-base font-semibold ${active ? 'text-navy' : 'text-text-secondary'}`}>{label}</span>
      <span
        className={`inline-flex items-center gap-1 text-xs font-medium ${
          complete ? 'text-emerald-600' : 'text-text-secondary'
        }`}
      >
        {complete && <CheckCircle2 size={13} />}
        {complete ? 'Complete' : 'Incomplete'}
      </span>
    </button>
  )
}

function OnboardingModal({ user, onUpdated, onCancel, mode = 'onboarding', onClose }) {
  const isEdit = mode === 'edit'
  const formDone = isFormComplete(user)
  const [activeTab, setActiveTab] = useState(isEdit ? 'form' : formDone ? 'shirt' : 'form')
  const [kotaAsal, setKotaAsal] = useState([])
  const [bandara, setBandara] = useState([])
  const [blazerSizes, setBlazerSizes] = useState([])
  const [submitting, setSubmitting] = useState(false)
  const [justSaved, setJustSaved] = useState(false)
  const savedTimeoutRef = useRef(null)

  useEffect(() => {
    getKotaAsalOptions().then((res) => setKotaAsal(res.data || []))
    getBandaraOptions().then((res) => setBandara(res.data || []))
    getBlazerSizeOptions().then((res) => setBlazerSizes(res.data || []))
    return () => clearTimeout(savedTimeoutRef.current)
  }, [])

  // Onboarding gets implicit feedback on save — the tab auto-advances, or
  // the whole modal disappears once the dashboard unlocks. Editing an
  // already-complete profile has neither, so without this a successful save
  // looks like clicking Save did nothing at all.
  const handleSaved = async () => {
    await onUpdated()
    setJustSaved(true)
    clearTimeout(savedTimeoutRef.current)
    savedTimeoutRef.current = setTimeout(() => setJustSaved(false), 2500)
  }

  useEffect(() => {
    // Auto-advance only applies to the mandatory onboarding flow (Form tab
    // flips to complete after a save) — editing an already-complete profile
    // should let the participant navigate the tabs freely, not get bounced.
    if (isEdit) return
    if (formDone && activeTab === 'form') setActiveTab('shirt')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formDone, isEdit])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 py-8 backdrop-blur-sm">
      <div className="flex max-h-[88vh] w-full max-w-2xl flex-col overflow-hidden rounded-2xl bg-white shadow-2xl">
        <div className="relative shrink-0 px-8 pt-8">
          {isEdit && (
            <button
              type="button"
              onClick={onClose}
              aria-label="Close"
              className="absolute right-6 top-6 text-text-secondary transition hover:text-text-primary"
            >
              <X size={20} />
            </button>
          )}
          <h2 className="text-center text-2xl font-bold text-navy">
            {isEdit ? 'Edit Data Diri' : 'Welcome to SCC 2026'}
            {user?.name ? `, ${user.name}` : ''}
          </h2>
          <p className="mt-1 text-center text-sm text-text-secondary">
            {isEdit ? 'Perbarui data formulir dan ukuran baju Anda.' : 'Please fill out these informations. Thank you!'}
          </p>

          <div className="mt-7 flex border-b border-surface-border">
            <TabHeader
              label="Form"
              complete={formDone}
              active={activeTab === 'form'}
              onClick={() => setActiveTab('form')}
            />
            <TabHeader
              label="Blazer Size"
              complete={isShirtComplete(user)}
              active={activeTab === 'shirt'}
              onClick={() => setActiveTab('shirt')}
            />
          </div>
          <div className="relative h-0.5 w-full bg-surface-border">
            <div
              className="absolute h-0.5 w-1/2 bg-gold transition-all duration-300"
              style={{ left: activeTab === 'form' ? '0%' : '50%' }}
            />
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-8 py-6">
          {activeTab === 'form' ? (
            <FormTab
              user={user}
              kotaAsal={kotaAsal}
              bandara={bandara}
              formId={FORM_ID}
              onSaved={handleSaved}
              onSubmittingChange={setSubmitting}
            />
          ) : (
            <ShirtSizeTab
              user={user}
              blazerSizes={blazerSizes}
              formId={FORM_ID}
              onSaved={handleSaved}
              onSubmittingChange={setSubmitting}
            />
          )}
        </div>

        <div className="flex shrink-0 flex-col gap-3 border-t border-surface-border bg-white px-8 py-5 shadow-[0_-6px_16px_-8px_rgba(11,24,48,0.15)]">
          {justSaved && (
            <p className="inline-flex items-center gap-1.5 text-xs font-semibold text-emerald-600">
              <CheckCircle2 size={14} />
              Perubahan tersimpan
            </p>
          )}
          <div className="flex items-center gap-3">
            <Button
              type="button"
              variant="secondary"
              className="flex-1"
              disabled={submitting}
              onClick={activeTab === 'shirt' ? () => setActiveTab('form') : isEdit ? onClose : onCancel}
            >
              {activeTab === 'shirt' ? 'Back' : isEdit ? 'Close' : 'Cancel'}
            </Button>
            <Button type="submit" form={FORM_ID} className="flex-1" disabled={submitting}>
              {submitting ? 'Menyimpan...' : isEdit ? 'Update' : 'Save'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default OnboardingModal
