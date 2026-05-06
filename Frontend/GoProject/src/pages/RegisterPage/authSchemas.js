import { z } from 'zod'

export const loginSchema = z.object({
  email: z.string().trim().email('Email không hợp lệ').max(255),
  password: z.string().min(1, 'Mật khẩu là bắt buộc'),
})

export const registerSchema = z.object({
  username: z.string().trim().min(3, 'Tên tối thiểu 3 ký tự').max(100),
  email: z.string().trim().email('Email không hợp lệ').max(255),
  password: z.string().min(8, 'Mật khẩu tối thiểu 8 ký tự').max(72),
  role: z.enum(['manager', 'member'], {
    errorMap: () => ({ message: 'Vai trò phải là manager hoặc member' }),
  }),
})
