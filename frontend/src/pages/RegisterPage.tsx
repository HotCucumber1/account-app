import {FormEvent, useState} from 'react'
import {Link, Navigate, useNavigate} from 'react-router-dom'
import {api} from '../api'
import {useAuth} from '../auth'

export function RegisterPage() {
    const {token} = useAuth()
    const navigate = useNavigate()
    const [name, setName] = useState('')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState<string | null>(null)
    const [pending, setPending] = useState(false)

    if (token) {
        return <Navigate to="/" replace/>
    }

    async function submit(event: FormEvent) {
        event.preventDefault()
        setPending(true)
        setError(null)
        try {
            await api.register(email, password, name)
            navigate('/login', {replace: true, state: {registered: true}})
        } catch (e) {
            setError(e instanceof Error ? e.message : 'Ошибка регистрации')
        } finally {
            setPending(false)
        }
    }

    return (
        <form className="card" onSubmit={submit}>
            <h1>Регистрация</h1>
            <label>
                Имя
                <input value={name} onChange={(e) => setName(e.target.value)} required maxLength={255}
                       autoComplete="name"/>
            </label>
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
                    autoComplete="new-password"
                />
            </label>
            {error && <p className="error">{error}</p>}
            <button type="submit" disabled={pending}>
                Зарегистрироваться
            </button>
            <p className="muted">
                Уже есть учётная запись? <Link to="/login">Войти</Link>
            </p>
        </form>
    )
}
