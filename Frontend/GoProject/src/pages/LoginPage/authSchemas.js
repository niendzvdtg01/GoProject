import { z } from 'zod'

export const loginSchema = z.object({
  email: z.string().trim().email('Email không hợp lệ').max(255),
  password: z.string().min(1, 'Mật khẩu là bắt buộc'),
})
