import {FormEvent, useState} from 'react'
import {Link} from 'react-router-dom'
import {api} from '../api'
import {useAuth} from '../auth'

export function CabinetPage() {
    const {token, account, setAccount, logout, handleError} = useAuth()
    const [editing, setEditing] = useState(false)
    const [name, setName] = useState(account?.name ?? '')
    const [error, setError] = useState<string | null>(null)
    const [saved, setSaved] = useState(false)
    const [pending, setPending] = useState(false)

    if (!account || !token) {
        return null
    }

    function startEditing() {
        setName(account!.name)
        setError(null)
        setSaved(false)
        setEditing(true)
    }

    async function save(event: FormEvent) {
        event.preventDefault()
        setPending(true)
        setError(null)
        try {
            const updated = await api.updateMe(token!, name)
            setAccount(updated)
            setEditing(false)
            setSaved(true)
        } catch (e) {
            handleError(e)
            setError(e instanceof Error ? e.message : 'Не удалось сохранить')
        } finally {
            setPending(false)
        }
    }

    return (
        <section className="card">
            <header className="row">
                <h1>Личный кабинет</h1>
                <button className="secondary" onClick={() => void logout()}>
                    Выйти
                </button>
            </header>

            {editing ? (
                <form onSubmit={save}>
                    <label>
                        Имя
                        <input value={name} onChange={(e) => setName(e.target.value)} required maxLength={255}
                               autoFocus/>
                    </label>
                    {error && <p className="error">{error}</p>}
                    <div className="row">
                        <button type="submit" disabled={pending}>
                            Сохранить
                        </button>
                        <button type="button" className="secondary" onClick={() => setEditing(false)}
                                disabled={pending}>
                            Отмена
                        </button>
                    </div>
                </form>
            ) : (
                <>
                    <dl>
                        <dt>Имя</dt>
                        <dd>{account.name}</dd>
                        <dt>Email</dt>
                        <dd>{account.email}</dd>
                        <dt>Зарегистрирован</dt>
                        <dd>{new Date(account.createdAt).toLocaleString('ru-RU')}</dd>
                    </dl>
                    {saved && <p className="success">Данные сохранены.</p>}
                    <div className="row">
                        <button onClick={startEditing}>Редактировать</button>
                        <Link className="button" to="/product">
                            Перейти в продукт
                        </Link>
                    </div>
                </>
            )}
        </section>
    )
}
