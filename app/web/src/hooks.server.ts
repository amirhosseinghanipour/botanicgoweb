import type { Handle } from '@sveltejs/kit';
import { paraglideMiddleware } from '$lib/paraglide/server';

const handleParaglide: Handle = ({ event, resolve }) => paraglideMiddleware(event.request, ({ request, locale }) => {
	event.request = request;

	return resolve(event, {
		transformPageChunk: ({ html }) => html.replace('%paraglide.lang%', locale)
	});
});

export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);

	// Set security headers
	response.headers.set('Content-Security-Policy', `
		default-src 'self';
		script-src 'self' 'unsafe-inline' 'unsafe-eval' ${process.env.NODE_ENV === 'development' ? 'ws: wss:' : ''};
		style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
		img-src 'self' data: blob: https: https://images.unsplash.com;
		font-src 'self' data: https://fonts.gstatic.com;
		connect-src 'self' http://localhost:8000 ${process.env.VITE_API_URL || 'http://localhost:8080'} ${process.env.NODE_ENV === 'development' ? 'ws: wss:' : ''} https://fonts.googleapis.com https://fonts.gstatic.com;
		media-src 'self' blob:;
		frame-ancestors 'none';
		form-action 'self';
		base-uri 'self';
		object-src 'none';
		worker-src 'self' blob:;
		${!process.env.VITE_API_URL?.includes('localhost') ? 'upgrade-insecure-requests;' : ''}
	`.replace(/\s+/g, ' ').trim());

	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('X-XSS-Protection', '1; mode=block');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');

	return response;
};
