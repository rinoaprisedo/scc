import { z } from 'zod'

export const loginSchema = z.object({
  email: z.string().email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
})

export const forgotPasswordSchema = z.object({
  email: z.string().email('Enter a valid email'),
})

export const resetPasswordSchema = z
  .object({
    password: z
      .string()
      .min(8, 'Min 8 characters')
      .regex(/[A-Z]/, 'Must contain an uppercase letter')
      .regex(/[0-9]/, 'Must contain a number')
      .regex(/[^A-Za-z0-9]/, 'Must contain a symbol'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

export const userSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Enter a valid email'),
  password: z.string().optional().or(z.literal('')),
  status: z.enum(['active', 'inactive', 'suspended']),
  role_uuid: z.string().optional().or(z.literal('')),
})

export const roleSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional().or(z.literal('')),
})

export const menuSectionSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  icon: z.string().optional().or(z.literal('')),
  order: z.union([z.string(), z.number()]).optional(),
})

export const menuSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  icon: z.string().optional().or(z.literal('')),
  path: z.string().min(1, 'Path is required'),
  menu_section_uuid: z.string().min(1, 'Section is required'),
  order: z.union([z.string(), z.number()]).optional(),
  is_active: z.boolean().optional(),
})

export const kotaAsalSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  province: z.string().optional().or(z.literal('')),
})

export const bandaraSchema = z.object({
  name: z.string().min(1, 'Name is required'),
})

export const qrGateSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  code: z
    .string()
    .min(1, 'Code is required')
    .regex(/^[A-Za-z0-9_-]+$/, 'Only letters, numbers, dash and underscore allowed'),
  points: z
    .union([z.string(), z.number()])
    .refine((v) => v !== '' && !Number.isNaN(Number(v)) && Number(v) >= 0, 'Points must be 0 or more'),
  is_reusable: z.boolean().optional(),
})

export const pesertaSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  // Email is optional — NIK (ktp_number) is the required identifier for
  // peserta accounts, which will eventually log in with NIK + password
  // instead of email.
  email: z.string().email('Enter a valid email').optional().or(z.literal('')),
  // Not shown on create (defaults to 'scc2026', set in PesertaFormModal) —
  // only editable when editing, where blank means keep the current one.
  password: z.string().optional().or(z.literal('')),
  status: z.enum(['active', 'inactive', 'suspended']),
  title: z.string().optional().or(z.literal('')),
  first_name: z.string().optional().or(z.literal('')),
  middle_name: z.string().optional().or(z.literal('')),
  last_name: z.string().optional().or(z.literal('')),
  birth_date: z.string().optional().or(z.literal('')),
  origin_city_uuid: z.string().optional().or(z.literal('')),
  origin_city_other: z.string().optional().or(z.literal('')),
  nearest_airport_uuid: z.string().optional().or(z.literal('')),
  dietary_restriction: z.string().optional().or(z.literal('')),
  dietary_restriction_other: z.string().optional().or(z.literal('')),
  phone_number: z.string().optional().or(z.literal('')),
  ktp_number: z.string().min(1, 'NIK is required'),
  nomor_ktp: z.string().optional().or(z.literal('')),
  passport_number: z.string().optional().or(z.literal('')),
  passport_expiry: z.string().optional().or(z.literal('')),
  jacket_size: z.string().optional().or(z.literal('')),
  polo_size: z.string().optional().or(z.literal('')),
  nomor_meja: z.string().optional().or(z.literal('')),
})
