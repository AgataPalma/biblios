import { describe, expect, it } from 'vitest'
import { DEFAULT_API_BASE_URL, resolveApiBaseUrl } from './api-url'

describe('resolveApiBaseUrl', () => {
    it('uses the same-origin API path when no URL is configured', () => {
        expect(resolveApiBaseUrl()).toBe(DEFAULT_API_BASE_URL)
        expect(resolveApiBaseUrl('   ')).toBe(DEFAULT_API_BASE_URL)
    })

    it('accepts root-relative and HTTPS API URLs', () => {
        expect(resolveApiBaseUrl('/custom/api')).toBe('/custom/api')
        expect(resolveApiBaseUrl('https://api.example.test/v1/')).toBe('https://api.example.test/v1')
    })

    it('rejects insecure, protocol-relative, malformed, and credential-bearing URLs', () => {
        expect(() => resolveApiBaseUrl('http://api.example.test/v1')).toThrow(/HTTPS/)
        expect(() => resolveApiBaseUrl('//api.example.test/v1')).toThrow(/root-relative/)
        expect(() => resolveApiBaseUrl('not a URL')).toThrow(/root-relative/)
        expect(() => resolveApiBaseUrl('https://user:pass@api.example.test/v1')).toThrow(/credentials/)
    })
})
