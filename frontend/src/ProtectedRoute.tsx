import {ReactNode} from 'react'
import {Navigate, useLocation} from 'react-router-dom'
import {useAuth} from './auth'

export function ProtectedRoute({children}: { children: ReactNode }) {
    const {token, account, loading, loadError} = useAuth()
    const location = useLocation()
    if (!token) {
        return <Navigate to="/login" replace state={{from: location.pathname}}/>
    }
    if (loading) {
        return <p className="muted">Загрузка…</p>
    }
    if (!account) {
        return <p className="error">{loadError ?? 'Не удалось загрузить профиль'}</p>
    }
    return <>{children}</>
}
