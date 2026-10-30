import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// API доступен только через gateway: /auth/... и /account/... проксируются на него.
const gateway = process.env.GATEWAY_URL ?? 'http://localhost:80'
const proxy = {
    '/auth/api': {target: gateway, changeOrigin: true},
    '/account/api': {target: gateway, changeOrigin: true},
}

export default defineConfig({
    plugins: [react()],
    server: {port: 5173, proxy},
    preview: {port: 4173, proxy},
})
