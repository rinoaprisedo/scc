import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import { useMutation } from '@tanstack/react-query'
import Input from '../../components/ui/Input'
import Button from '../../components/ui/Button'
import { login, me } from '../../api/auth'
import { loginSchema } from '../../utils/validation'
import useAuthStore from '../../store/authStore'

function Login() {
  const navigate = useNavigate()
  const setUser = useAuthStore((s) => s.setUser)
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({ resolver: zodResolver(loginSchema) })

  const mutation = useMutation({
    mutationFn: login,
    onSuccess: async () => {
      try {
        const res = await me()
        setUser(res.data)
      } catch {
        // ignore; guard will redirect if session invalid
      }
      toast.success('Welcome back!')
      navigate('/')
    },
    onError: (err) => {
      toast.error(err.response?.data?.message || 'Login failed')
    },
  })

  return (
    <div>
      <h2 className="mb-4 text-center text-lg font-semibold text-text-primary">
        Sign in to your account
      </h2>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit((data) => mutation.mutate(data))}>
        <Input label="Email" type="email" placeholder="you@example.com" error={errors.email?.message} {...register('email')} />
        <Input label="Password" type="password" placeholder="••••••••" error={errors.password?.message} {...register('password')} />
        <Button type="submit" disabled={mutation.isPending} className="w-full">
          {mutation.isPending ? 'Signing in...' : 'Sign in'}
        </Button>
      </form>
    </div>
  )
}

export default Login
