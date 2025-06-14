import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { api } from '$lib/api/client';

export const load: PageServerLoad = async ({ params }) => {
  try {
    const session = await api.getSession(params.sessionId);
    return { session };
  } catch (err) {
    console.error('Failed to load chat session:', err);
    throw error(404, 'Chat session not found');
  }
}; 