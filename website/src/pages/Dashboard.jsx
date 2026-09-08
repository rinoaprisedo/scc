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
import { getMyScans, getLeaderboard } from '../api/qrGate'
import { getActiveSliders } from '../api/sliders'
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
import Carousel from '../components/ui/Carousel'
import { isFormComplete } from '../utils/onboarding'
import { isPastDeadline } from '../utils/deadline'
import { fileURL } from '../utils/url'
import bgDesktop from '../assets/Microsite-02.webp'
import bgMobile from '../assets/Microsite-01 (1).webp'
import danamonLogo from '../assets/Single Logo_Logo White.webp'
import eventBadge from '../assets/Microsite-02.png'
import lpsLogo from '../assets/logo-lps.avif'
import bottomStrip from '../assets/Danamon List-05.jpg'
import profileBanner from '../assets/KV SCC 2026_Artboard 5.webp'
import leaderboardBanner from '../assets/KV SCC 2026_Artboard 3 copy.webp'

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

// Ranked list of participants by total QR gate points — rendered in place
// of the slider carousel in both layout slots (desktop TopRankCard box +
// mobile banner) whenever the admin's "Leaderboard" website-menu toggle is
// on (see settings.menu_leaderboard_enabled).
function LeaderboardCard({ entries, className = '' }) {
  const rankBadgeClass = (rank) => {
    if (rank === 1) return 'bg-gold text-navy-dark'
    if (rank === 2) return 'bg-white/70 text-navy-dark'
    if (rank === 3) return 'bg-amber-700 text-white'
    return 'bg-white/10 text-white/70'
  }

  return (
    <div className={`flex flex-col gap-3 rounded-2xl border-2 border-gold-light/60 bg-navy p-5 ${className}`}>
      <div className="flex items-center gap-2 text-gold">
        <Trophy size={18} />
        <p className="text-base font-bold text-white">Leaderboard</p>
      </div>
      {entries.length === 0 ? (
        <p className="py-6 text-center text-sm text-white/50">Belum ada data</p>
      ) : (
        <ol className="flex flex-col gap-2">
          {entries.map((entry) => (
            <li
              key={entry.uuid}
              className="flex items-center justify-between gap-3 rounded-xl bg-navy-light/60 px-3 py-2"
            >
              <div className="flex min-w-0 items-center gap-3">
                <span
                  className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold ${rankBadgeClass(entry.rank)}`}
                >
                  {entry.rank}
                </span>
                <span className="truncate text-sm font-medium text-white">{entry.name}</span>
              </div>
              <span className="shrink-0 text-sm font-bold text-gold">{entry.points} Poin</span>
            </li>
          ))}
        </ol>
      )}
    </div>
  )
}

// Banner carousel slot — images come from the admin-managed Sliders module
// (filtered to this slot's type), falling back to a static asset when
// nothing has been uploaded/activated yet so the box never renders empty.
function SliderCard({ sliders, fallback, fallbackAlt, className = '' }) {
  const images =
    sliders.length > 0 ? sliders.map((s) => ({ src: fileURL(s.image), alt: 'Slider' })) : [{ src: fallback, alt: fallbackAlt }]
  return <Carousel images={images} className={className} />
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
  const [sliders, setSliders] = useState([])
  const [leaderboard, setLeaderboard] = useState([])

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
    getActiveSliders()
      .then((res) => setSliders(res.data || []))
      .catch(() => {})
    getLeaderboard(10)
      .then((res) => setLeaderboard(res.data || []))
      .catch(() => {})
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
  const leaderboardEnabled = publicSettings.menu_leaderboard_enabled === 'true'
  const desktopSliders = sliders.filter((s) => s.type === 'desktop')
  const mobileSliders = sliders.filter((s) => s.type === 'mobile')

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

      <header className="relative flex flex-col items-start gap-4 px-6 py-6 sm:px-10 sm:py-8 md:grid md:grid-cols-3 md:items-center md:gap-4">
        <div className="flex w-full items-start justify-between md:contents">
          <img
            src={danamonLogo}
            alt="Danamon — A member of MUFG"
            className="h-[70px] md:h-[120px] md:justify-self-start lg:h-[144px]"
          />
          {stage === 'done' && (
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-orange-500 to-red-500 px-4 py-2 text-sm font-semibold text-white shadow-md transition hover:opacity-90 md:hidden"
            >
              <LogOut size={14} />
              Logout
            </button>
          )}
        </div>
        <img
          src={eventBadge}
          alt="Sales Champion Conference 2026 — The Art of Leadership"
          className="h-[220px] self-center md:h-[164px] md:justify-self-center lg:h-[230px]"
        />
        {stage === 'done' && (
          <button
            type="button"
            onClick={handleLogout}
            className="hidden items-center gap-2 rounded-lg bg-gradient-to-r from-orange-500 to-red-500 px-5 py-2.5 text-base font-semibold text-white shadow-md transition hover:opacity-90 md:inline-flex md:justify-self-end"
          >
            <LogOut size={16} />
            Logout
          </button>
        )}
      </header>

      {stage === 'done' && (
        <div className="relative mx-auto flex w-full max-w-[1440px] flex-col gap-6 px-6 pb-10 sm:px-10 lg:flex-row">
          <div className="order-last hidden w-full md:block lg:order-none lg:w-[34rem] lg:shrink-0">
            {leaderboardEnabled ? (
              <LeaderboardCard entries={leaderboard} />
            ) : (
              <SliderCard
                sliders={desktopSliders}
                fallback={leaderboardBanner}
                fallbackAlt="Top Rank Group"
                className="border-2 border-gold-light/60"
              />
            )}
          </div>

          <div className="flex flex-1 flex-col gap-6">
            {leaderboardEnabled ? (
              <LeaderboardCard entries={leaderboard} className="shadow-lg md:hidden" />
            ) : (
              <SliderCard
                sliders={mobileSliders}
                fallback={profileBanner}
                fallbackAlt=""
                className="border-2 border-gold-light/60 shadow-lg md:hidden"
              />
            )}

            <div className="relative flex flex-col items-center gap-6 overflow-hidden rounded-2xl border-2 border-gold-light/60 bg-navy p-7 text-center shadow-lg sm:flex-row sm:items-center sm:justify-between sm:p-9 sm:text-left">
              <div className="relative flex flex-col items-center gap-3 sm:flex-row sm:items-center sm:gap-5">
                <div className="hidden h-16 w-16 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-gold-light via-gold to-gold-dark text-2xl font-bold text-navy-dark shadow-md ring-4 ring-gold/20 sm:flex">
                  {user?.name?.charAt(0).toUpperCase() || 'P'}
                </div>
                <div>
                  <p className="mt-1 text-2xl font-bold text-white">{user?.name || 'Participant'}</p>
                  <p className="mt-1 text-base font-medium text-white/60">NIP {user?.ktp_number || '-'}</p>
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
                    className={`group relative flex cursor-pointer items-center justify-between overflow-hidden rounded-2xl border-2 border-gold-light/60 bg-navy px-8 py-7 shadow-lg transition hover:border-transparent hover:bg-gradient-to-br hover:from-gold-light hover:via-gold hover:to-gold-dark ${
                      // QRIS Cross Border becomes a floating "Upload QR" button on
                      // mobile (below) instead of a grid tile — see the fixed
                      // button after this grid.
                      action === 'qris' ? 'hidden md:flex' : ''
                    }`}
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

      {stage === 'done' && (
        <div className="fixed bottom-5 left-1/2 z-40 flex -translate-x-1/2 flex-col items-center gap-1.5 md:hidden">
          <button
            type="button"
            aria-label="Upload QR"
            onClick={() => {
              if (!isMenuEnabled(publicSettings, 'menu_qris_cross_border_enabled')) {
                setAlertMessage('Maaf, Fitur ini belum tersedia')
                return
              }
              setQrisOpen(true)
            }}
            className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-gradient-to-r from-gold-light via-gold to-gold-dark text-navy-dark shadow-xl transition hover:opacity-90"
          >
            <QrCode size={26} />
          </button>
          <span className="rounded-full bg-navy-dark/80 px-2.5 py-0.5 text-[11px] font-bold text-white shadow">
            Upload QR
          </span>
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
