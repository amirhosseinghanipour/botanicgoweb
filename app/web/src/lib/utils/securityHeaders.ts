import { browser } from '$app/environment';
import { dev } from '$app/environment';

// Get API URL based on environment
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const isLocalhost = API_URL.includes('localhost') || API_URL.includes('127.0.0.1');

// Default Content Security Policy
const DEFAULT_CSP = `
    default-src 'self';
    script-src 'self' 'unsafe-inline' 'unsafe-eval' ${dev ? 'ws: wss:' : ''};
    style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
    img-src 'self' data: blob: https: https://images.unsplash.com;
    font-src 'self' data: https://fonts.gstatic.com;
    connect-src 'self' ${API_URL} ${dev ? 'ws: wss:' : ''} https://fonts.googleapis.com https://fonts.gstatic.com;
    media-src 'self' blob:;
    frame-ancestors 'none';
    form-action 'self';
    base-uri 'self';
    object-src 'none';
    worker-src 'self' blob:;
    ${!isLocalhost ? 'upgrade-insecure-requests;' : ''}
`.replace(/\s+/g, ' ').trim();

// Set security headers using meta tags
export function setSecurityHeaders(): void {
    if (!browser) return;

    // Content Security Policy
    const cspMeta = document.createElement('meta');
    cspMeta.httpEquiv = 'Content-Security-Policy';
    cspMeta.content = DEFAULT_CSP;
    document.head.appendChild(cspMeta);

    // X-Content-Type-Options
    const contentTypeMeta = document.createElement('meta');
    contentTypeMeta.httpEquiv = 'X-Content-Type-Options';
    contentTypeMeta.content = 'nosniff';
    document.head.appendChild(contentTypeMeta);

    // X-Frame-Options
    const frameOptionsMeta = document.createElement('meta');
    frameOptionsMeta.httpEquiv = 'X-Frame-Options';
    frameOptionsMeta.content = 'DENY';
    document.head.appendChild(frameOptionsMeta);

    // X-XSS-Protection
    const xssMeta = document.createElement('meta');
    xssMeta.httpEquiv = 'X-XSS-Protection';
    xssMeta.content = '1; mode=block';
    document.head.appendChild(xssMeta);

    // Referrer-Policy
    const referrerMeta = document.createElement('meta');
    referrerMeta.httpEquiv = 'Referrer-Policy';
    referrerMeta.content = 'strict-origin-when-cross-origin';
    document.head.appendChild(referrerMeta);

    // Permissions-Policy
    const permissionsMeta = document.createElement('meta');
    permissionsMeta.httpEquiv = 'Permissions-Policy';
    permissionsMeta.content = 'camera=(), microphone=(), geolocation=()';
    document.head.appendChild(permissionsMeta);
}

// Update CSP dynamically
export function updateCSP(directives: Record<string, string[]>): void {
    if (!browser) return;

    const cspMeta = document.querySelector('meta[http-equiv="Content-Security-Policy"]');
    if (!cspMeta) return;

    const newDirectives = Object.entries(directives)
        .map(([key, values]) => `${key} ${values.join(' ')}`)
        .join('; ');

    cspMeta.setAttribute('content', newDirectives);
}

// Add trusted domain to CSP
export function addTrustedDomain(domain: string): void {
    if (!browser) return;

    const cspMeta = document.querySelector('meta[http-equiv="Content-Security-Policy"]');
    if (!cspMeta) return;

    const currentCSP = cspMeta.getAttribute('content') || '';
    const newCSP = currentCSP.replace(
        /connect-src ([^;]+)/,
        (match, p1) => `connect-src ${p1} ${domain}`
    );

    cspMeta.setAttribute('content', newCSP);
}

// Remove trusted domain from CSP
export function removeTrustedDomain(domain: string): void {
    if (!browser) return;

    const cspMeta = document.querySelector('meta[http-equiv="Content-Security-Policy"]');
    if (!cspMeta) return;

    const currentCSP = cspMeta.getAttribute('content') || '';
    const newCSP = currentCSP.replace(
        new RegExp(`connect-src ([^;]+) ${domain}`),
        (match, p1) => `connect-src ${p1}`
    );

    cspMeta.setAttribute('content', newCSP);
} 