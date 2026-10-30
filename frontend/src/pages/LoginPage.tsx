import {FormEvent, useState} from 'react'
import {Link, Navigate, useLocation, useNavigate} from 'react-router-dom'
import {useAuth} from '../auth'

export function LoginPage() {
    const {token, login} = useAuth()
    const navigate = useNavigate()
    const location = useLocation()
    const from = (location.state as { from?: string; registered?: boolean } | null) ?? {}
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState<string | null>(null)
    const [pending, setPending] = useState(false)

    if (token) {
        return <Navigate to={from.from ?? '/'} replace/>
    }

    async function submit(event: FormEvent) {
        event.preventDefault()
        setPending(true)
        setError(null)
        try {
            await login(email, password)
            navigate(from.from ?? '/', {replace: true})
        } catch (e) {
            setError(e instanceof Error ? e.message : 'Ошибка входа')
        } finally {
            setPending(false)
        }
    }

    return (
        <form className="card" onSubmit={submit}>
            <h1>Вход</h1>
            {from.registered && <p className="success">Регистрация прошла успешно, войдите.</p>}
            <label>
                Email
                <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required
                       autoComplete="email"/>
            </label>
            <label>
                Пароль
                <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    autoComplete="current-password"
                />
            </label>
            {error && <p className="error">{error}</p>}
            <button type="submit" disabled={pending}>
                Войти
            </button>
            <p className="muted">
                Нет учётной записи? <Link to="/register">Зарегистрироваться</Link>
            </p>
        </form>
    )
}
