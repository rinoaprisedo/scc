import { useEffect, useState, useMemo, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { UploadCloud, X } from 'lucide-react'
import Input from '../../components/ui/Input'
import Button from '../../components/ui/Button'
import Skeleton from '../../components/ui/Skeleton'
import FileUpload from '../../components/ui/FileUpload'
import { Menubar, MenubarLabel } from '../../components/ui/Menubar'
import { getSettings, updateSettings, uploadSetting, uploadSettingImages, removeSettingImage } from '../../api/settings'
import { fileURL } from '../../utils/url'
import { cn } from '../../utils/cn'
import usePageActions from '../../hooks/usePageActions'
import useSettingsStore from '../../store/settingsStore'

const tabs = ['General', 'Appearance', 'Website Content', 'Website Menu', 'Registration', 'Maintenance']

// agenda_file/dress_code_file stay single-file (image or PDF).
const SINGLE_CONTENT_ITEMS = [
  { key: 'agenda_file', label: 'Agenda Acara' },
  { key: 'dress_code_file', label: 'Dress Code' },
]

// event_information_file/about_malaysia_file are multi-image — the website
// renders them as a Carousel slider in the preview popup once more than one
// image exists.
const MULTI_CONTENT_ITEMS = [
  { key: 'event_information_file', label: 'Event Information' },
  { key: 'about_malaysia_file', label: 'About Malaysia' },
]

// Mirrors backend settings.parseImagePaths — a value that isn't valid JSON
// is a pre-multi-image legacy single path, not a parse failure.
function parseImagePaths(raw) {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : [raw]
  } catch {
    return [raw]
  }
}

function MultiImageUpload({ target, label, paths, uploading, onUploaded, onRemoved }) {
  const inputRef = useRef(null)
  const [isDragging, setIsDragging] = useState(false)

  const handleFiles = (files) => {
    const list = Array.from(files || [])
    if (list.length > 0) onUploaded(list)
  }

  return (
    <div className="flex flex-col gap-2">
      <p className="text-xs font-medium uppercase tracking-[.05em] text-text-secondary">{label}</p>

      {paths.length > 0 && (
        <div className="grid grid-cols-4 gap-2">
          {paths.map((path) => (
            <div
              key={path}
              className="group relative overflow-hidden rounded-md border border-surface-border bg-surface-card"
            >
              <img src={fileURL(path)} alt={label} className="h-20 w-full object-cover" />
              <button
                type="button"
                onClick={() => onRemoved(path)}
                className="absolute right-1 top-1 rounded-full bg-black/60 p-1 text-white opacity-0 transition-opacity group-hover:opacity-100"
                aria-label="Remove image"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
          ))}
        </div>
      )}

      <div
        role="button"
        tabIndex={0}
        onClick={() => inputRef.current?.click()}
        onKeyDown={(e) => e.key === 'Enter' && inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault()
          setIsDragging(true)
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={(e) => {
          e.preventDefault()
          setIsDragging(false)
          handleFiles(e.dataTransfer.files)
        }}
        className={cn(
          'flex cursor-pointer items-center gap-3 rounded-lg border-2 border-dashed border-surface-border bg-white px-4 py-3 text-sm text-text-secondary transition-colors duration-150 hover:border-primary/60 hover:bg-primary/5',
          isDragging && 'border-primary bg-primary/5',
        )}
      >
        <input
          ref={inputRef}
          type="file"
          accept="image/png,image/jpeg,image/webp"
          multiple
          className="hidden"
          onChange={(e) => {
            handleFiles(e.target.files)
            e.target.value = ''
          }}
        />
        <UploadCloud className="h-4 w-4 shrink-0 text-primary" />
        <span>{uploading ? 'Uploading...' : 'Click or drag to add image(s)'}</span>
      </div>
    </div>
  )
}

// Mirrors the participant website's Dashboard MENU_ITEMS — toggling one off
// makes the site show a "belum tersedia" popup instead of opening it.
const MENU_TOGGLES = [
  { key: 'menu_event_agenda_enabled', label: 'Event Agenda' },
  { key: 'menu_event_gallery_enabled', label: 'Event Gallery' },
  { key: 'menu_dress_code_enabled', label: 'Dress Code' },
  { key: 'menu_qris_cross_border_enabled', label: 'QRIS Cross Border' },
  { key: 'menu_about_malaysia_enabled', label: 'About Malaysia' },
  { key: 'menu_scanner_qr_enabled', label: 'Scanner QR' },
  { key: 'menu_history_scanner_enabled', label: 'History Scanner' },
  { key: 'menu_event_information_enabled', label: 'Event Information' },
]

// Not a home-tile toggle like MENU_TOGGLES above — checking this swaps the
// participant website's slider carousel slots for real leaderboard content.
// Kept in its own list so it can render with different copy/default.
const LEADERBOARD_TOGGLE = { key: 'menu_leaderboard_enabled', label: 'Leaderboard' }

function isPdf(path) {
  return !!path && path.toLowerCase().endsWith('.pdf')
}

function Settings() {
  const [tab, setTab] = useState('General')
  const [appName, setAppName] = useState('')
  const [maintenance, setMaintenance] = useState(false)
  const [logoPreview, setLogoPreview] = useState(null)
  const [faviconPreview, setFaviconPreview] = useState(null)
  const [primaryColor, setPrimaryColorInput] = useState('#c2622e')
  const [menuToggles, setMenuToggles] = useState({})
  const [registrationDeadline, setRegistrationDeadline] = useState('')
  const [formEditDeadline, setFormEditDeadline] = useState('')
  const queryClient = useQueryClient()
  const applyPrimaryColorLive = useSettingsStore((s) => s.setPrimaryColor)
  const storedPrimaryColor = useSettingsStore((s) => s.primaryColor)

  const { data, isLoading } = useQuery({ queryKey: ['settings'], queryFn: getSettings })

  useEffect(() => {
    if (data?.data) {
      // GET /settings returns a { key: value } map directly, not an array.
      const map = data.data
      setAppName(map.app_name || '')
      setMaintenance(map.maintenance_mode === 'true' || map.maintenance_mode === true)
      setLogoPreview(fileURL(map.app_logo))
      setFaviconPreview(fileURL(map.app_favicon))
      setPrimaryColorInput(map.primary_color || storedPrimaryColor)
      const toggles = {}
      MENU_TOGGLES.forEach(({ key }) => {
        toggles[key] = map[key] !== 'false' // missing key defaults to enabled
      })
      // Opposite default from MENU_TOGGLES — a missing/unseeded key means
      // "off", not "on", since this replaces content rather than hiding a tile.
      toggles[LEADERBOARD_TOGGLE.key] = map[LEADERBOARD_TOGGLE.key] === 'true'
      setMenuToggles(toggles)
      setRegistrationDeadline(map.registration_deadline || '')
      setFormEditDeadline(map.form_edit_deadline || '')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  const handlePrimaryColorChange = (hex) => {
    setPrimaryColorInput(hex)
    applyPrimaryColorLive(hex) // live-preview across the whole app before saving
  }

  const saveMutation = useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      toast.success('Settings saved')
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Save failed'),
  })

  const uploadMutation = useMutation({
    mutationFn: uploadSetting,
    onSuccess: () => {
      toast.success('File uploaded')
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Upload failed'),
  })

  const uploadImagesMutation = useMutation({
    mutationFn: ({ target, files }) => uploadSettingImages(target, files),
    onSuccess: () => {
      toast.success('Images uploaded')
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Upload failed'),
  })

  const removeImageMutation = useMutation({
    mutationFn: ({ target, path }) => removeSettingImage(target, path),
    onSuccess: () => {
      toast.success('Image removed')
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Remove failed'),
  })

  const handleMenuToggleChange = (key) => (e) => {
    setMenuToggles((prev) => ({ ...prev, [key]: e.target.checked }))
  }

  const handleSaveMenuToggles = () => {
    const payload = {}
    MENU_TOGGLES.forEach(({ key }) => {
      payload[key] = menuToggles[key] ? 'true' : 'false'
    })
    payload[LEADERBOARD_TOGGLE.key] = menuToggles[LEADERBOARD_TOGGLE.key] ? 'true' : 'false'
    saveMutation.mutate(payload)
  }

  const handleContentUpload = (key) => (file) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('target', key)
    uploadMutation.mutate(formData)
  }

  const handleFileChange = (key) => (file) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('target', key)
    uploadMutation.mutate(formData)
    const reader = new FileReader()
    reader.onload = () => (key === 'app_logo' ? setLogoPreview(reader.result) : setFaviconPreview(reader.result))
    reader.readAsDataURL(file)
  }

  usePageActions(
    useMemo(
      () => (
        <Menubar>
          <MenubarLabel>Settings</MenubarLabel>
        </Menubar>
      ),
      [],
    ),
  )

  if (isLoading) return <Skeleton className="h-96 w-full" />

  return (
    <div className="space-y-4">
      <div className="flex gap-1 border-b border-surface-border">
        {tabs.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium ${
              tab === t ? 'border-b-2 border-primary text-primary' : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      <div className="rounded-lg border border-surface-border bg-surface-card p-6">
        {tab === 'General' && (
          <div className="max-w-md space-y-4">
            <Input label="App Name" value={appName} onChange={(e) => setAppName(e.target.value)} />
            <Button onClick={() => saveMutation.mutate({ app_name: appName })} disabled={saveMutation.isPending}>
              Save
            </Button>
          </div>
        )}

        {tab === 'Appearance' && (
          <div className="max-w-md space-y-6">
            <FileUpload
              label="Logo"
              hint="PNG, JPEG or WebP"
              accept="image/png,image/jpeg,image/webp"
              preview={logoPreview}
              onFileSelect={handleFileChange('app_logo')}
              previewClassName="h-16 w-16"
            />
            <FileUpload
              label="Favicon"
              hint="PNG, JPEG or WebP"
              accept="image/png,image/jpeg,image/webp"
              preview={faviconPreview}
              onFileSelect={handleFileChange('app_favicon')}
              previewClassName="h-12 w-12"
            />
            <div>
              <p className="mb-2 text-sm font-medium text-text-primary">Base Color</p>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={primaryColor}
                  onChange={(e) => handlePrimaryColorChange(e.target.value)}
                  className="h-10 w-14 cursor-pointer rounded-md border border-surface-border bg-transparent p-1"
                />
                <Input
                  value={primaryColor}
                  onChange={(e) => handlePrimaryColorChange(e.target.value)}
                  className="w-32 uppercase"
                  maxLength={7}
                />
              </div>
              <p className="mt-1 text-xs text-text-secondary">Applied instantly as a preview; click Save to persist.</p>
              <Button
                className="mt-3"
                onClick={() => saveMutation.mutate({ primary_color: primaryColor })}
                disabled={saveMutation.isPending}
              >
                Save Color
              </Button>
            </div>
          </div>
        )}

        {tab === 'Website Content' && (
          <div className="max-w-md space-y-6">
            <p className="text-sm text-text-secondary">
              Image or PDF shown in the participant website's Agenda Acara / Dress Code popups, and one or more
              images for the Event Information / About Malaysia popups (rendered as a slider once more than one is
              added).
            </p>
            {SINGLE_CONTENT_ITEMS.map(({ key, label }) => {
              const path = data?.data?.[key]
              const pdf = isPdf(path)
              return (
                <div key={key}>
                  <FileUpload
                    label={label}
                    hint="PNG, JPEG, WebP, or PDF"
                    accept="image/png,image/jpeg,image/webp,application/pdf"
                    preview={path && !pdf ? fileURL(path) : null}
                    onFileSelect={handleContentUpload(key)}
                    previewClassName="h-16 w-16"
                  />
                  {path && (
                    <a
                      href={fileURL(path)}
                      target="_blank"
                      rel="noreferrer"
                      className="mt-1 inline-block text-xs font-medium text-primary hover:underline"
                    >
                      {pdf ? 'View current PDF' : 'View current file'}
                    </a>
                  )}
                </div>
              )
            })}
            {MULTI_CONTENT_ITEMS.map(({ key, label }) => (
              <MultiImageUpload
                key={key}
                target={key}
                label={label}
                paths={parseImagePaths(data?.data?.[key])}
                uploading={uploadImagesMutation.isPending && uploadImagesMutation.variables?.target === key}
                onUploaded={(files) => uploadImagesMutation.mutate({ target: key, files })}
                onRemoved={(path) => removeImageMutation.mutate({ target: key, path })}
              />
            ))}
          </div>
        )}

        {tab === 'Website Menu' && (
          <div className="max-w-md space-y-4">
            <p className="text-sm text-text-secondary">
              Turn a tile off to show "Maaf, Fitur ini belum tersedia" instead of opening it on the participant
              website.
            </p>
            <div className="space-y-3">
              {MENU_TOGGLES.map(({ key, label }) => (
                <label key={key} className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    checked={menuToggles[key] ?? true}
                    onChange={handleMenuToggleChange(key)}
                    className="h-4 w-4 accent-primary"
                  />
                  <span className="text-sm text-text-primary">{label}</span>
                </label>
              ))}
            </div>

            <div className="border-t border-surface-border pt-4">
              <p className="mb-2 text-sm text-text-secondary">
                When checked, the slider carousel on the participant website's home page (both the desktop and
                mobile slots) is replaced with the real-time leaderboard instead.
              </p>
              <label className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={menuToggles[LEADERBOARD_TOGGLE.key] ?? false}
                  onChange={handleMenuToggleChange(LEADERBOARD_TOGGLE.key)}
                  className="h-4 w-4 accent-primary"
                />
                <span className="text-sm text-text-primary">{LEADERBOARD_TOGGLE.label}</span>
              </label>
            </div>

            <Button onClick={handleSaveMenuToggles} disabled={saveMutation.isPending}>
              Save
            </Button>
          </div>
        )}

        {tab === 'Registration' && (
          <div className="max-w-md space-y-6">
            <div>
              <p className="mb-1 text-sm font-medium text-text-primary">Attendance Confirmation Deadline</p>
              <p className="mb-2 text-xs text-text-secondary">
                After this, a participant who hasn't answered the attendance prompt yet sees "Maaf, registrasi sudah
                ditutup" on login instead. Anyone who already answered (hadir or tidak hadir) is unaffected. Leave
                blank for no deadline.
              </p>
              <input
                type="datetime-local"
                value={registrationDeadline}
                onChange={(e) => setRegistrationDeadline(e.target.value)}
                className="h-[42px] w-full rounded-md border-[1.5px] border-surface-border bg-white px-3 text-sm text-text-primary outline-none focus:border-primary"
              />
              <Button
                className="mt-3"
                onClick={() => saveMutation.mutate({ registration_deadline: registrationDeadline })}
                disabled={saveMutation.isPending}
              >
                Save
              </Button>
            </div>

            <div>
              <p className="mb-1 text-sm font-medium text-text-primary">Form Edit Deadline</p>
              <p className="mb-2 text-xs text-text-secondary">
                After this, participants can no longer submit or edit their profile form — they see a warning
                instead. Leave blank for no deadline.
              </p>
              <input
                type="datetime-local"
                value={formEditDeadline}
                onChange={(e) => setFormEditDeadline(e.target.value)}
                className="h-[42px] w-full rounded-md border-[1.5px] border-surface-border bg-white px-3 text-sm text-text-primary outline-none focus:border-primary"
              />
              <Button
                className="mt-3"
                onClick={() => saveMutation.mutate({ form_edit_deadline: formEditDeadline })}
                disabled={saveMutation.isPending}
              >
                Save
              </Button>
            </div>
          </div>
        )}

        {tab === 'Maintenance' && (
          <div className="max-w-md space-y-4">
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={maintenance}
                onChange={(e) => setMaintenance(e.target.checked)}
                className="h-4 w-4 accent-primary"
              />
              <span className="text-sm text-text-primary">Enable maintenance mode</span>
            </label>
            <Button
              onClick={() => saveMutation.mutate({ maintenance_mode: maintenance })}
              disabled={saveMutation.isPending}
            >
              Save
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}

export default Settings
