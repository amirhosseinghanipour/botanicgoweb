import { API_URL } from '$lib/config';

export async function exchangeCodeForToken(provider: string, code: string, state: string) {
  const response = await fetch(`${API_URL}/api/auth/${provider}/callback`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ code, state })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to exchange code for token');
  }

  return response.json();
} 