export type Account = {
    id: string
    email: string
    name: string
    createdAt: string
}

export type Tokens = {
    accessToken: string
    accessTokenExpiresAt: string
    accountID: string
    sessionID: string
}

export class ApiError extends Error {
    constructor(
        readonly status: number,
        message: string,
    ) {
        super(message)
    }
}

async function request<T>(method: string, path: string, token?: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = {}
    if (body !== undefined) {
        headers['Content-Type'] = 'application/json'
    }
    if (token) {
        headers.Authorization = `Bearer ${token}`
    }
    const response = await fetch(path, {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
    })
    const text = await response.text()
    const data = text ? safeParse(text) : undefined
    if (!response.ok) {
        const message = (data as { message?: string } | undefined)?.message ?? (text.trim() || response.statusText)
        throw new ApiError(response.status, message)
    }
    return data as T
}

function safeParse(text: string): unknown {
    try {
        return JSON.parse(text)
    } catch {
        return undefined
    }
}

export const api = {
    register: (email: string, password: string, name: string) =>
        request<{ accountID: string }>('POST', '/auth/api/public/v1/auth/register', undefined, {email, password, name}),
    login: (email: string, password: string) =>
        request<Tokens>('POST', '/auth/api/public/v1/auth/login', undefined, {email, password}),
    logout: (token: string) => request<void>('POST', '/auth/api/public/v1/auth/logout', token),
    getMe: (token: string) => request<Account>('GET', '/account/api/public/v1/accounts/me', token),
    updateMe: (token: string, name: string) =>
        request<Account>('PATCH', '/account/api/public/v1/accounts/me', token, {name}),
}
