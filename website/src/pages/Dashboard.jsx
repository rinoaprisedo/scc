import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Calendar,
  Image,
  Shirt,
  ListChecks,
  Info,
  User,
  FileEdit,
  Trophy,
  LogOut,
  ChevronRight,
  History,
  Newspaper,
  QrCode,
} from 'lucide-react'
import { getMe, logout } from '../api/auth'
import { getPublicSettings } from '../api/settings'
import { getMyScans } from '../api/qrGate'
import AttendanceModal from '../components/onboarding/AttendanceModal'
import DeclinedScreen from '../components/onboarding/DeclinedScreen'
import NoticeScreen from '../components/onboarding/NoticeScreen'
import OnboardingModal from '../components/onboarding/OnboardingModal'
import ProfileModal from '../components/ProfileModal'
import ContentPreviewModal from '../components/ContentPreviewModal'
import AlertModal from '../components/AlertModal'
import ScannerModal from '../components/ScannerModal'
import HistoryScannerModal from '../components/HistoryScannerModal'
import QrisCrossBorderModal from '../components/QrisCrossBorderModal'
import { isFormComplete } from '../utils/onboarding'
import { isPastDeadline } from '../utils/deadline'
import bgDesktop from '../assets/Microsite-02.webp'
import bgMobile from '../assets/Microsite-01 (1).webp'
import danamonLogo from '../assets/Single Logo_Logo White.webp'
import eventBadge from '../assets/Microsite-02.png'
import lpsLogo from '../assets/logo-lps.avif'
import bottomStrip from '../assets/Danamon List-05.jpg'

// attendance_status is the single source of truth from the backend
// (belum_konfirmasi / hadir_belum_lengkap / tidak_hadir / hadir_lengkap) —
// isFormComplete/isShirtComplete are only needed here to pick WHICH tab to
// land on while status is "hadir_belum_lengkap".
function onboardingStage(user) {
  if (!user) return 'done'
  switch (user.attendance_status) {
    case 'tidak_hadir':
      return 'declined'
    case 'hadir_lengkap':
      return 'done'
    case 'hadir_belum_lengkap':
      return isFormComplete(user) ? 'shirt' : 'form'
    case 'belum_konfirmasi':
    default:
      return 'attendance'
  }
}

// Agenda/Dress Code open a preview popup for a file the admin
// uploads under that settings key (image or PDF); the rest are static tiles
// until their own detail page ships.
const MENU_ITEMS = [
  { label: 'Event Agenda', icon: Calendar, contentKey: 'agenda_file', settingKey: 'menu_event_agenda_enabled' },
  {
    label: 'Event Information',
    icon: Newspaper,
    contentKey: 'event_information_file',
    settingKey: 'menu_event_information_enabled',
  },
  { label: 'Dress Code', icon: Shirt, contentKey: 'dress_code_file', settingKey: 'menu_dress_code_enabled' },
  { label: 'QRIS Cross Border', icon: ListChecks, action: 'qris', settingKey: 'menu_qris_cross_border_enabled' },
  {
    label: 'About Malaysia',
    icon: Info,
    contentKey: 'about_malaysia_file',
    settingKey: 'menu_about_malaysia_enabled',
  },
  { label: 'Event Gallery', icon: Image, settingKey: 'menu_event_gallery_enabled' },
  { label: 'Scan QR', icon: QrCode, action: 'scan', settingKey: 'menu_scanner_qr_enabled' },
  { label: 'History Point', icon: History, action: 'history', settingKey: 'menu_history_scanner_enabled' },
]

// A missing key means the admin never toggled it — treat as enabled so
// existing tiles keep working until someone explicitly turns one off.
function isMenuEnabled(publicSettings, settingKey) {
  return publicSettings[settingKey] !== 'false'
}

const PODIUM = [
  { place: 2, size: 'h-14 w-14', ring: 'ring-slate-300/40' },
  { place: 1, size: 'h-16 w-16', ring: 'ring-gold/40' },
  { place: 3, size: 'h-14 w-14', ring: 'ring-orange-300/40' },
]

function TopRankCard() {
  return (
    <div className="order-last flex w-full flex-col gap-7 rounded-2xl border-2 border-gold-light/60 bg-navy p-9 shadow-lg lg:order-none lg:w-96 lg:shrink-0">
      <div className="flex items-center gap-2">
        <Trophy size={20} className="text-gold" />
        <p className="text-lg font-bold text-white">Top Rank Group</p>
      </div>

      <div className="flex items-end justify-center gap-4">
        {PODIUM.map(({ place, size, ring }) => (
          <div key={place} className="flex flex-col items-center gap-2">
            <div
              className={`flex ${size} items-center justify-center rounded-full border-2 border-dashed border-gold/30 bg-navy-light text-white/40 ring-4 ${ring}`}
            >
              <Trophy size={place === 1 ? 26 : 20} />
            </div>
            <span className="flex h-5 w-5 items-center justify-center rounded-full bg-navy-light text-[10px] font-bold text-white/70">
              {place}
            </span>
          </div>
        ))}
      </div>

      <div className="flex flex-col gap-2.5">
        {[4, 5, 6].map((rank) => (
          <div key={rank} className="flex items-center gap-3 rounded-lg bg-navy-light px-3 py-2.5">
            <span className="text-xs font-semibold text-white/60">{rank}</span>
            <div className="h-2.5 flex-1 rounded-full bg-white/10" />
            <div className="h-2.5 w-10 rounded-full bg-white/10" />
          </div>
        ))}
      </div>

      <p className="text-center text-xs font-medium text-white/70">Rankings coming soon</p>
    </div>
  )
}

function Dashboard() {
  const navigate = useNavigate()
  const [user, setUser] = useState(null)
  const [checking, setChecking] = useState(true)
  const [editingProfile, setEditingProfile] = useState(false)
  const [viewingProfile, setViewingProfile] = useState(false)
  const [publicSettings, setPublicSettings] = useState({})
  const [preview, setPreview] = useState(null)
  const [alertMessage, setAlertMessage] = useState(null)
  const [scannerOpen, setScannerOpen] = useState(false)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [qrisOpen, setQrisOpen] = useState(false)
  const [totalPoints, setTotalPoints] = useState(0)

  const refreshUser = () => getMe().then((res) => setUser(res.data))
  const refreshPoints = () => getMyScans().then((res) => setTotalPoints(res.data?.total_points || 0))

  useEffect(() => {
    refreshUser()
      .catch(() => navigate('/'))
      .finally(() => setChecking(false))
    getPublicSettings()
      .then((res) => setPublicSettings(res.data || {}))
      .catch(() => {})
    refreshPoints().catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigate])

  const handleLogout = async () => {
    try {
      await logout()
    } finally {
      navigate('/')
    }
  }

  if (checking) {
    return <div className="flex min-h-screen items-center justify-center bg-navy text-white">Loading...</div>
  }

  const stage = onboardingStage(user)
  const registrationClosed = isPastDeadline(publicSettings.registration_deadline)
  const formEditClosed = isPastDeadline(publicSettings.form_edit_deadline)

  const handleOpenEditForm = () => {
    if (formEditClosed) {
      setAlertMessage('Maaf, batas waktu pengisian/edit formulir sudah lewat')
      return
    }
    setEditingProfile(true)
  }

  return (
    <div className="relative flex min-h-screen flex-col overflow-hidden bg-navy">
      <img src={bgMobile} alt="" className="fixed inset-0 h-full w-full object-cover md:hidden" />
      <img src={bgDesktop} alt="" className="fixed inset-0 hidden h-full w-full object-cover md:block" />

      <header className="relative flex flex-col items-start gap-4 px-6 py-6 sm:px-10 sm:py-8 md:flex-row md:items-start md:justify-between">
        <img src={danamonLogo} alt="Danamon — A member of MUFG" className="h-[70px] md:h-[120px] md:translate-y-8 lg:h-[144px] lg:translate-y-10" />
        <img
          src={eventBadge}
          alt="Sales Champion Conference 2026 — The Art of Leadership"
          className="h-[220px] self-center md:h-[164px] md:self-auto lg:h-[230px]"
        />
      </header>

      {stage === 'done' && (
        <div className="relative mx-auto flex w-full max-w-[1440px] flex-col gap-6 px-6 pb-10 sm:px-10 lg:flex-row">
          <TopRankCard />

          <div className="flex flex-1 flex-col gap-6">
            <div className="relative flex flex-col items-center gap-6 overflow-hidden rounded-2xl border-2 border-gold-light/60 bg-navy p-7 text-center shadow-lg sm:flex-row sm:items-center sm:justify-between sm:p-9 sm:text-left">
              <div className="relative flex flex-col items-center gap-3 sm:flex-row sm:items-center sm:gap-5">
                <div className="hidden h-16 w-16 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-gold-light via-gold to-gold-dark text-2xl font-bold text-navy-dark shadow-md ring-4 ring-gold/20 sm:flex">
                  {user?.name?.charAt(0).toUpperCase() || 'P'}
                </div>
                <div>
                  <p className="mt-1 text-2xl font-bold text-white">{user?.name || 'Participant'}</p>
                  <p className="mt-1 text-base font-medium text-white/60">NIK {user?.ktp_number || '-'}</p>
                  <p className="mt-2 inline-flex items-center gap-1.5 rounded-full bg-gold/10 px-3 py-1 text-sm font-bold text-gold">
                    <Trophy size={14} />
                    {totalPoints} Poin
                  </p>
                </div>
              </div>

              <div className="relative grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:flex-wrap">
                <button
                  type="button"
                  onClick={() => setViewingProfile(true)}
                  className="inline-flex items-center justify-center gap-1.5 rounded-full bg-navy-light px-4 py-2 text-sm font-semibold text-white transition hover:bg-white/10 sm:w-auto"
                >
                  <User size={14} />
                  View Profile
                </button>
                <button
                  type="button"
                  onClick={handleOpenEditForm}
                  className="inline-flex items-center justify-center gap-1.5 rounded-full bg-navy-light px-4 py-2 text-sm font-semibold text-white transition hover:bg-white/10 sm:w-auto"
                >
                  <FileEdit size={14} />
                  Edit Form
                </button>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="col-span-2 inline-flex items-center justify-center gap-1.5 rounded-full bg-gradient-to-r from-orange-500 to-red-500 px-4 py-2 text-sm font-semibold text-white shadow-md transition hover:opacity-90 sm:col-span-1 sm:w-auto"
                >
                  <LogOut size={14} />
                  Logout
                </button>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
              {MENU_ITEMS.map(({ label, icon: Icon, contentKey, action, settingKey }) => {
                const enabled = isMenuEnabled(publicSettings, settingKey)
                const handleActivate = () => {
                  if (!enabled) {
                    setAlertMessage('Maaf, Fitur ini belum tersedia')
                    return
                  }
                  if (contentKey) {
                    setPreview({ title: label, path: publicSettings[contentKey] })
                  } else if (action === 'scan') {
                    setScannerOpen(true)
                  } else if (action === 'history') {
                    setHistoryOpen(true)
                  } else if (action === 'qris') {
                    setQrisOpen(true)
                  }
                }
                return (
                  <div
                    key={label}
                    role="button"
                    tabIndex={0}
                    onClick={handleActivate}
                    onKeyDown={(e) => e.key === 'Enter' && handleActivate()}
                    className="group relative flex cursor-pointer items-center justify-between overflow-hidden rounded-2xl border-2 border-gold-light/60 bg-navy px-8 py-7 shadow-lg transition hover:border-transparent hover:bg-gradient-to-br hover:from-gold-light hover:via-gold hover:to-gold-dark"
                  >
                    <div className="pointer-events-none absolute -bottom-10 -right-10 h-32 w-56 rotate-12 bg-gradient-to-tr from-transparent via-gold/25 to-transparent blur-2xl transition group-hover:opacity-0" />

                    <div className="relative flex items-center gap-4">
                      <Icon size={26} strokeWidth={1.7} className="shrink-0 text-gold transition group-hover:text-white" />
                      <span className="text-2xl font-bold text-white">{label}</span>
                    </div>
                    <ChevronRight size={20} className="relative shrink-0 text-gold transition group-hover:text-white" />
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      )}

      <footer className="relative mt-auto pt-2 sm:pt-4">
        <div className="px-6 pt-6 sm:px-10 sm:pt-8">
          <p className="text-sm font-semibold text-white sm:text-base">Bancassurance</p>
          <p className="mt-1 flex flex-wrap items-center gap-2 text-[10px] font-medium text-white/80 sm:text-xs">
            <span>
              PT Bank Danamon Indonesia Tbk berizin dan diawasi oleh Otoritas Jasa Keuangan dan Bank
              Indonesia serta merupakan peserta penjaminan LPS
            </span>
            <img src={lpsLogo} alt="LPS" className="h-5 w-auto" />
          </p>
        </div>
        <img src={bottomStrip} alt="" className="mt-6 h-3 w-full object-cover sm:h-4" />
      </footer>

      {preview && <ContentPreviewModal title={preview.title} path={preview.path} onClose={() => setPreview(null)} />}

      {alertMessage && <AlertModal message={alertMessage} onClose={() => setAlertMessage(null)} />}

      {scannerOpen && (
        <ScannerModal
          onClose={() => {
            setScannerOpen(false)
            refreshPoints().catch(() => {})
          }}
        />
      )}

      {historyOpen && <HistoryScannerModal onClose={() => setHistoryOpen(false)} />}

      {qrisOpen && <QrisCrossBorderModal onClose={() => setQrisOpen(false)} />}

      {viewingProfile && (
        <ProfileModal
          user={user}
          onClose={() => setViewingProfile(false)}
          onEdit={() => {
            setViewingProfile(false)
            handleOpenEditForm()
          }}
        />
      )}

      {stage === 'attendance' &&
        (registrationClosed ? (
          <NoticeScreen title="Maaf" message="Registrasi sudah ditutup." onLogout={handleLogout} />
        ) : (
          <AttendanceModal userName={user?.name} onUpdated={refreshUser} onLogout={handleLogout} />
        ))}

      {stage === 'declined' && (
        <DeclinedScreen
          onLogout={handleLogout}
          onUpdated={refreshUser}
          registrationClosed={registrationClosed}
        />
      )}

      {(stage === 'form' || stage === 'shirt') && formEditClosed ? (
        <NoticeScreen title="Maaf" message="Batas waktu pengisian formulir sudah lewat." onLogout={handleLogout} />
      ) : (
        (stage === 'form' || stage === 'shirt' || editingProfile) && (
          <OnboardingModal
            user={user}
            onUpdated={refreshUser}
            onCancel={handleLogout}
            mode={stage === 'done' ? 'edit' : 'onboarding'}
            onClose={() => setEditingProfile(false)}
          />
        )
      )}
    </div>
  )
}

export default Dashboard
