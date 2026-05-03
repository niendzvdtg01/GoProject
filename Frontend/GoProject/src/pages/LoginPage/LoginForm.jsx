import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { ROUTES } from '../../shared/constants/routes.js'
import { Button } from '../../shared/components/Button.jsx'
import { Input } from '../../shared/components/Input.jsx'
import { getApiErrorMessage } from '../../shared/services/apiError.js'
import { useLogin } from './useLogin.js'
import { loginSchema } from './authSchemas.js'

export function LoginForm() {
  const mutation = useLogin()
  const {
    formState: { errors },
    handleSubmit,
    register,
  } = useForm({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  })

  return (
    <form className="grid gap-4 p-8" onSubmit={handleSubmit((values) => mutation.mutate(values))}>
      <div>
        <h2 className="text-xl font-extrabold text-slate-950">Đăng nhập</h2>
        <p className="mt-1 text-sm text-slate-600">Sử dụng JWT access token từ backend Go.</p>
      </div>

      <Input label="Email" type="email" error={errors.email?.message} {...register('email')} />
      <Input label="Mật khẩu" type="password" error={errors.password?.message} {...register('password')} />

      {mutation.isError ? (
        <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm font-semibold text-red-700">
          {getApiErrorMessage(mutation.error)}
        </p>
      ) : null}

      <Button type="submit" isLoading={mutation.isPending}>
        Đăng nhập
      </Button>
      <Link className="text-center text-sm font-bold text-sky-700 hover:text-sky-900" to={ROUTES.register}>
        Chưa có tài khoản? Đăng ký
      </Link>
    </form>
  )
}
