export const DEFAULT_API_BASE_URL = '/api/v1'

export function resolveApiBaseUrl(configuredUrl?: string): string {
    const value = configuredUrl?.trim() || DEFAULT_API_BASE_URL

    if (value.startsWith('/') && !value.startsWith('//')) {
        return value
    }

    let parsed: URL
    try {
        parsed = new URL(value)
    } catch {
        throw new Error('VITE_API_URL must be a root-relative path or an absolute HTTPS URL')
    }

    if (parsed.protocol !== 'https:') {
        throw new Error('Absolute VITE_API_URL values must use HTTPS')
    }
    if (parsed.username || parsed.password) {
        throw new Error('VITE_API_URL must not include credentials')
    }

    return value.replace(/\/$/, '')
}
