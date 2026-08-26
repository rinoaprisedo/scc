import Button from './ui/Button'

function AgreementModal({ onAgree }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 shadow-2xl">
        <h2 className="text-2xl font-bold text-navy">Persetujuan Data Pribadi</h2>

        <p className="mt-4 text-sm leading-relaxed text-text-secondary">
          Dengan melanjutkan, saya menyatakan telah memahami dan menyetujui bahwa data pribadi yang
          saya berikan dapat dikumpulkan, digunakan, dan dibagikan kepada pihak ketiga yang terkait
          dengan pelaksanaan Sales Champion Conference (SCC) 2026, termasuk namun tidak terbatas
          pada penyedia jasa perjalanan, akomodasi, transportasi, serta penyedia perlengkapan
          peserta.
        </p>

        <p className="mt-4 text-sm leading-relaxed text-text-secondary">
          Data pribadi tersebut hanya akan digunakan untuk keperluan penyelenggaraan dan
          pelaksanaan SCC 2026 sesuai dengan ketentuan perlindungan data pribadi yang berlaku dan
          tidak digunakan untuk tujuan lain di luar keperluan tersebut.
        </p>

        <Button className="mt-7 w-full" onClick={onAgree}>
          Saya Setuju dan Lanjutkan
        </Button>
      </div>
    </div>
  )
}

export default AgreementModal
