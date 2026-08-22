import { z } from 'zod'

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
    dietary_restriction_other: z.string().optional().or(z.literal('')),
    phone_number: z.string().min(1, 'Wajib diisi').regex(/^\d+$/, 'Hanya angka'),
    nomor_ktp: z
      .string()
      .regex(/^\d*$/, 'Hanya angka')
      .optional()
      .or(z.literal('')),
    passport_number: z.string().min(1, 'Wajib diisi'),
    passport_expiry: z.string().min(1, 'Wajib diisi'),
  })
  .superRefine((val, ctx) => {
    if (val.origin_city_uuid === 'other' && !val.origin_city_other) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['origin_city_other'], message: 'Wajib diisi' })
    }
    if (val.dietary_restriction === 'Other' && !val.dietary_restriction_other) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['dietary_restriction_other'], message: 'Wajib diisi' })
    }
  })

export const shirtSchema = z.object({
  jacket_size: z.string().min(1, 'Wajib dipilih'),
  polo_size: z.string().min(1, 'Wajib dipilih'),
})
