import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import LoginModal from '../components/LoginModal.jsx'
import bgMobile from '../assets/Microsite-04 (1).jpg'
import bgDesktop from '../assets/Microsite-03.jpg'
import badge from '../assets/Microsite-08.png'
import title from '../assets/Microsite-10.png'
import divider from '../assets/Microsite-03.png'
import danamonLogo from '../assets/logo-danamon.webp'
import lpsLogo from '../assets/logo-lps.avif'

function Landing() {
  const [showLogin, setShowLogin] = useState(false)
  const navigate = useNavigate()

  return (
    <div className="relative flex min-h-screen flex-col overflow-hidden bg-navy">
      <img src={bgMobile} alt="" className="absolute inset-0 h-full w-full object-cover md:hidden" />
      <img src={bgDesktop} alt="" className="absolute inset-0 hidden h-full w-full object-cover md:block" />

      <header className="relative px-6 py-6 sm:px-10 sm:py-8">
        <img src={danamonLogo} alt="Danamon — A member of MUFG" className="h-[96px] sm:h-[116px]" />
      </header>

      <div className="relative grid flex-1 grid-cols-1 items-center gap-2 px-6 py-8 md:grid-cols-2 md:gap-8 md:px-16">
        <div className="flex items-center justify-center">
          <img
            src={badge}
            alt="Sales Champion Conference 2026 — Brilliance Beyond Boundaries, A Decade of Champions"
            className="-mt-8 w-full max-w-[584px] md:-mt-16 md:max-w-[648px]"
          />
        </div>

        <div className="flex flex-col items-center justify-center gap-2 md:gap-8">
          <img
            src={title}
            alt="Brilliance Beyond Boundaries — A Decade of Champions"
            className="w-full max-w-[220px] md:max-w-sm"
          />
          <img src={divider} alt="" className="w-full max-w-[200px] md:max-w-xs" />

          <button
            type="button"
            onClick={() => setShowLogin(true)}
            className="mt-4 rounded-lg bg-gradient-to-r from-gold-light via-gold to-gold-dark px-10 py-3 text-sm font-bold uppercase tracking-widest text-navy-dark shadow-lg transition hover:opacity-90"
          >
            Login
          </button>
        </div>
      </div>

      <footer className="relative">
        <div className="px-6 py-6 sm:px-10 sm:py-8">
          <p className="text-sm font-semibold text-white sm:text-base">Bancassurance</p>
          <p className="mt-1 flex flex-wrap items-center gap-2 text-[10px] font-medium text-white/80 sm:text-xs">
            <span>
              PT Bank Danamon Indonesia Tbk berizin dan diawasi oleh Otoritas Jasa Keuangan dan Bank
              Indonesia serta merupakan peserta penjaminan LPS
            </span>
            <img src={lpsLogo} alt="LPS" className="h-5 w-auto" />
          </p>
        </div>
        {/* flat 3-band strip — CSS gradient instead of a raster image so it never shows browser downscale artifacts at this height */}
        <div
          className="mt-0.5 h-3 w-full sm:h-4"
          style={{
            background:
              'linear-gradient(to right, #FCBF28 0%, #FCBF28 21.7%, #F05A27 21.7%, #F05A27 70.7%, #F89422 70.7%, #F89422 100%)',
          }}
        />
      </footer>

      {showLogin && (
        <LoginModal onClose={() => setShowLogin(false)} onLoginSuccess={() => navigate('/dashboard')} />
      )}
    </div>
  )
}

export default Landing
