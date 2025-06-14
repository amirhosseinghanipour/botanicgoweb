import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { api } from '$lib/api/client';
import { goto } from '$app/navigation';
import type { User } from '$lib/types';
import { secureStorage } from '$lib/utils/secureStorage';
import { API_URL } from '$lib/config';

interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    error: string | null;
}

// Helper function to get cookie value
function getCookie(name: string): string | null {
    if (!browser) return null;
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return match ? match[2] : null;
}

const createAuthStore = () => {
    const { subscribe, set, update } = writable<AuthState>({
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: true,
        error: null
    });

    // Initialize auth state from storage
    const initializeAuth = async () => {
        if (!browser) return;

        console.log('Initializing auth store...');
        setLoading(true);

        try {
            const userData = localStorage.getItem('user');
            const token = await secureStorage.getSecureItem('token');
            const expiration = await secureStorage.getSecureItem('tokenExpiration');
            const sessionId = await secureStorage.getSecureItem('sessionId');

            console.log('Auth initialization state:', {
                hasUserData: !!userData,
                hasToken: !!token,
                hasExpiration: !!expiration,
                hasSessionId: !!sessionId,
                userData: userData ? JSON.parse(userData) : null,
                token: token ? 'present' : null,
                expiration: expiration,
                sessionId: sessionId
            });

            if (!userData || !token || !expiration) {
                console.log('Missing auth data, clearing state');
                set({
                    user: null,
                    token: null,
                    isAuthenticated: false,
                    isLoading: false,
                    error: null
                });
                return;
            }

            const expirationDate = new Date(expiration);
            if (expirationDate < new Date()) {
                console.log('Token expired, clearing state');
                set({
                    user: null,
                    token: null,
                    isAuthenticated: false,
                    isLoading: false,
                    error: null
                });
                return;
            }

            console.log('Verifying token with backend...');
            const response = await fetch(`${API_URL}/api/auth/verify`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                console.log('Token verification failed, clearing state');
                set({
                    user: null,
                    token: null,
                    isAuthenticated: false,
                    isLoading: false,
                    error: null
                });
                return;
            }

            const user = JSON.parse(userData);
            set({
                user,
                token,
                isAuthenticated: true,
                isLoading: false,
                error: null
            });
            // Initialize API client with token and session
            await api.initializeToken(token, parseInt(expiration, 10), sessionId || undefined);
            console.log('Auth initialized successfully with valid token');
        } catch (error) {
            console.error('Error initializing auth:', error);
            set({
                user: null,
                token: null,
                isAuthenticated: false,
                isLoading: false,
                error: null
            });
        }
    };

    const checkAuth = async () => {
        if (!browser) return false;

        try {
            const userData = localStorage.getItem('user');
            const token = await secureStorage.getSecureItem('token');
            const expiration = await secureStorage.getSecureItem('tokenExpiration');
            const sessionId = await secureStorage.getSecureItem('sessionId');

            if (!userData || !token || !expiration) {
                return false;
            }

            const expirationDate = new Date(expiration);
            if (expirationDate < new Date()) {
                return false;
            }

            const response = await fetch(`${API_URL}/api/auth/verify`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                return false;
            }

            const user = JSON.parse(userData);
            set({
                user,
                token,
                isAuthenticated: true,
                isLoading: false,
                error: null
            });
            // Initialize API client with token and session
            await api.initializeToken(token, parseInt(expiration, 10), sessionId || undefined);

            return true;
        } catch (error) {
            console.error('Error checking auth:', error);
            return false;
        }
    };

    const setLoading = (loading: boolean) => {
        update((state: AuthState) => ({ ...state, isLoading: loading }));
    };

    const setAuth = async (user: User, token: string, sessionId?: string) => {
        if (browser) {
            // Set loading state first
            setLoading(true);
            
            try {
                // Store user data in localStorage
                localStorage.setItem('user', JSON.stringify(user));
                // Store token in secureStorage
                const expirationTime = Date.now() + (30 * 24 * 60 * 60 * 1000); // 30 days
                await secureStorage.setSecureItem('token', token);
                await secureStorage.setSecureItem('tokenExpiration', expirationTime.toString());
                if (sessionId) {
                    await secureStorage.setSecureItem('sessionId', sessionId);
                }
                
                // Initialize API client with token
                await api.initializeToken(token, expirationTime, sessionId);
                
                // Add a small delay to ensure smooth transition
                await new Promise(resolve => setTimeout(resolve, 300));
                
                set({
                    user,
                    token,
                    isAuthenticated: true,
                    isLoading: false,
                    error: null
                });
            } catch (error) {
                console.error('Error setting auth state:', error);
                setLoading(false);
                throw error;
            }
        } else {
            set({
                user,
                token,
                isAuthenticated: true,
                isLoading: false,
                error: null
            });
        }
    };

    const clearAuth = async () => {
        if (browser) {
            // Set loading state first
            setLoading(true);
            
            try {
                // Add a small delay to ensure smooth transition
                await new Promise(resolve => setTimeout(resolve, 300));
                
                localStorage.removeItem('user');
                await secureStorage.removeSecureItem('token');
                await secureStorage.removeSecureItem('tokenExpiration');
                await secureStorage.removeSecureItem('sessionId');
                
                set({
                    user: null,
                    token: null,
                    isAuthenticated: false,
                    isLoading: false,
                    error: null
                });
            } catch (error) {
                console.error('Error clearing auth state:', error);
                setLoading(false);
                throw error;
            }
        } else {
            set({
                user: null,
                token: null,
                isAuthenticated: false,
                isLoading: false,
                error: null
            });
        }
    };

    const setError = (error: string) => {
        update(state => ({ ...state, error, isLoading: false }));
    };

    const login = async (email: string, password: string, rememberMe: boolean = false) => {
        setLoading(true);
        try {
            const response = await api.login({ email, password, rememberMe });
            setAuth(response.user, response.token);
            return response.user;
        } catch (err) {
            const error = err instanceof Error ? err.message : 'Failed to login';
            setError(error);
            throw err;
        }
    };

    const register = async (email: string, password: string) => {
        setLoading(true);
        try {
            const response = await api.register({ email, password });
            setAuth(response.user, response.token);
            return response.user;
        } catch (err) {
            const error = err instanceof Error ? err.message : 'Failed to register';
            setError(error);
            throw err;
        }
    };

    const logout = async () => {
        try {
            await api.logout();
        } finally {
            clearAuth();
            goto('/');
        }
    };

    return {
        subscribe,
        set: (state: AuthState) => set(state),
        setUser: (user: User | null) => update(state => ({ ...state, user })),
        setToken: (token: string | null) => update(state => ({ ...state, token })),
        setLoading,
        setError,
        initializeAuth,
        checkAuth,
        setAuth,
        clearAuth,
        login,
        register,
        logout
    };
};

export const auth = createAuthStore(); 