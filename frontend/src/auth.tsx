import {createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState} from 'react'
import {Account, api, ApiError} from './api'

const tokenKey = 'accessToken'

type AuthState = {
    token: string | null
    account: Account | null
    loading: boolean
    loadError: string | null
    login: (email: string, password: string) => Promise<void>
    logout: () => Promise<void>
    setAccount: (account: Account) => void
    handleError: (error: unknown) => void
}

const AuthContext = createContext<AuthState | null>(null)

function readToken(): string | null {
    try {
        return localStorage.getItem(tokenKey)
    } catch {
        return null
    }
}

function writeToken(token: string | null) {
    try {
        if (token) {
            localStorage.setItem(tokenKey, token)
        } else {
            localStorage.removeItem(tokenKey)
        }
    } catch {
        // Хранилище недоступно: токен живёт только в памяти вкладки.
    }
}

export function AuthProvider({children}: { children: ReactNode }) {
    const [token, setToken] = useState<string | null>(readToken)
    const [account, setAccount] = useState<Account | null>(null)
    const [loading, setLoading] = useState<boolean>(token !== null)
    const [loadError, setLoadError] = useState<string | null>(null)

    const reset = useCallback(() => {
        writeToken(null)
        setToken(null)
        setAccount(null)
    }, [])

    const handleError = useCallback(
        (error: unknown) => {
            if (error instanceof ApiError && error.status === 401) {
                reset()
            }
        },
        [reset],
    )

    useEffect(() => {
        if (!token) {
            setLoading(false)
            return
        }
        let cancelled = false
        setLoading(true)
        setLoadError(null)
        api
            .getMe(token)
            .then((me) => {
                if (!cancelled) setAccount(me)
            })
            .catch((error: unknown) => {
                if (cancelled) return
                if (error instanceof ApiError && error.status === 401) {
                    reset()
                } else {
                    setLoadError(error instanceof Error ? error.message : 'Не удалось загрузить профиль')
                }
            })
            .finally(() => {
                if (!cancelled) setLoading(false)
            })
        return () => {
            cancelled = true
        }
    }, [token, reset])

    const login = useCallback(async (email: string, password: string) => {
        const tokens = await api.login(email, password)
        writeToken(tokens.accessToken)
        setToken(tokens.accessToken)
    }, [])

    const logout = useCallback(async () => {
        if (token) {
            await api.logout(token).catch(() => undefined)
        }
        reset()
    }, [token, reset])

    const value = useMemo(
        () => ({token, account, loading, loadError, login, logout, setAccount, handleError}),
        [token, account, loading, loadError, login, logout, handleError],
    )
    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
    const context = useContext(AuthContext)
    if (!context) {
        throw new Error('useAuth must be used inside AuthProvider')
    }
    return context
}
