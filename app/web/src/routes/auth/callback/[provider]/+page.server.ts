import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { API_URL } from '$lib/config';

export const load: PageServerLoad = async ({ params, url, fetch, cookies }) => {
    const provider = params.provider;
    const code = url.searchParams.get('code');
    const state = url.searchParams.get('state');
    const oauthError = url.searchParams.get('error');

    if (oauthError) {
        throw error(400, `OAuth error: ${oauthError}`);
    }

    if (!code || !state) {
        throw error(400, 'Missing code or state parameter');
    }

    try {
        // Make a request to our backend API to handle the OAuth callback
        const response = await fetch(`${API_URL}/api/auth/${provider}/callback${url.search}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            credentials: 'include',
            mode: 'cors'
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw error(response.status, errorData.message || 'Authentication failed');
        }

        // Get the Set-Cookie headers from the response
        const setCookieHeaders = response.headers.getSetCookie();
        
        // Set the cookies in SvelteKit's cookie store
        for (const cookie of setCookieHeaders) {
            const [cookieStr] = cookie.split(';');
            const [name, value] = cookieStr.split('=');
            cookies.set(name, value, {
                path: '/',
                httpOnly: true,
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax'
            });
        }

        // Get the response data
        const data = await response.json();
        
        // Store user data in a cookie
        cookies.set('user_data', JSON.stringify(data.user), {
            path: '/',
            httpOnly: true,
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'lax',
            maxAge: 86400 // 24 hours
        });

        // Redirect to the complete page
        throw redirect(303, '/auth/callback/complete');
    } catch (err: any) {
        console.error('OAuth callback error:', err);
        if (err instanceof Response) throw err;
        if (err.status === 303) throw err; // Re-throw redirects
        throw error(500, 'Failed to complete authentication');
    }
};