import { z } from 'zod'

// Passport must stay valid at least 6 months past the event date
// (2026-10-10) — standard international travel-document requirement.
export const MIN_PASSPORT_EXPIRY = '2027-04-10'

// Every Form-tab field is required except middle_name (many participants
// don't have one) — mirrors the admin's peserta schema shape but stricter,
// since the website enforces completion before letting a participant into
// the dashboard while the admin panel allows saving partial records.
export const formSchema = z
  .object({
    title: z.string().min(1, 'Wajib dipilih'),
    first_name: z.string().min(1, 'Wajib diisi'),
    middle_name: z.string().optional().or(z.literal('')),
    last_name: z.string().min(1, 'Wajib diisi'),
    birth_date: z.string().min(1, 'Wajib diisi'),
    origin_city_uuid: z.string().min(1, 'Wajib dipilih'),
    origin_city_other: z.string().optional().or(z.literal('')),
    nearest_airport_uuid: z.string().min(1, 'Wajib dipilih'),
    dietary_restriction: z.string().min(1, 'Wajib dipilih'),
    phone_number: z.string().min(1, 'Wajib diisi').regex(/^\d+$/, 'Hanya angka'),
    nomor_ktp: z
      .string()
      .min(1, 'Wajib diisi')
      .regex(/^\d{16}$/, 'Nomor KTP harus 16 digit angka'),
    passport_number: z
      .string()
      .min(1, 'Wajib diisi')
      .min(8, 'Nomor paspor minimal 8 karakter')
      .max(9, 'Nomor paspor maksimal 9 karakter')
      .regex(/^[A-Za-z0-9]+$/, 'Nomor paspor hanya boleh huruf dan angka')
      .refine(
        (val) => /[A-Za-z]/.test(val) && /\d/.test(val),
        'Nomor paspor harus kombinasi huruf dan angka',
      ),
    passport_expiry: z
      .string()
      .min(1, 'Wajib diisi')
      .refine(
        (val) => val >= MIN_PASSPORT_EXPIRY,
        'Masa berlaku paspor minimal 6 bulan setelah 10 Oktober 2026 (hingga 10 April 2027)',
      ),
  })
  .superRefine((val, ctx) => {
    if (val.origin_city_uuid === 'other' && !val.origin_city_other) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['origin_city_other'], message: 'Wajib diisi' })
    }
  })

export const shirtSchema = z.object({
  blazer_size: z.string().min(1, 'Wajib dipilih'),
})
