import { browser } from '$app/environment';

const CSRF_TOKEN_KEY = 'csrf_token';

export function getCsrfToken(): string | null {
    if (!browser) return null;
    return localStorage.getItem(CSRF_TOKEN_KEY);
}

export function setCsrfToken(token: string): void {
    if (!browser) return;
    localStorage.setItem(CSRF_TOKEN_KEY, token);
}

export function removeCsrfToken(): void {
    if (!browser) return;
    localStorage.removeItem(CSRF_TOKEN_KEY);
}

export function generateCsrfToken(): string {
    return crypto.randomUUID();
} 