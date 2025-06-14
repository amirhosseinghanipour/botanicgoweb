import { browser } from '$app/environment';
import { NetworkError, ErrorCodes, type ErrorCode } from './errors';
import { getCsrfToken, setCsrfToken, removeCsrfToken, generateCsrfToken } from '$lib/utils/csrf';
import { rateLimiter } from '$lib/utils/rateLimiter';
import { secureStorage } from '$lib/utils/secureStorage';
import { goto } from '$app/navigation';
import type { Message } from '$lib/types';
import { API_URL } from '$lib/config';
import { auth } from '$lib/stores/auth';

interface User {
    id: string;
    email: string;
    name?: string;
    avatarUrl?: string;
    preferences: {
        theme: string;
        language: string;
        timezone: string;
        notifications: boolean;
    };
    provider: string;
    providerId?: string;
    createdAt: string;
    updatedAt: string;
}

interface AuthResponse {
    user: User;
    token: string;
    session: {
        id: string;
        expires_at: string;
    };
}

interface ChatSession {
    id: string;
    user_id: string;
    title: string;
    model: string;
    created_at: string;
    updated_at: string;
    messages: Message[];
}

interface UpdateProfileRequest {
    name?: string;
    avatarUrl?: string;
    preferences?: {
        theme: string;
    };
}

interface UpdatePreferencesRequest {
    theme?: string;
    language?: string;
    timezone?: string;
    notifications?: boolean;
}

export interface SessionInfo {
    id: string;
    createdAt: string;
    expiresAt: string;
    device: string;
    location: string;
}

interface LoginRequest {
    email: string;
    password: string;
    rememberMe?: boolean;
}

interface RetryConfig {
    maxRetries: number;
    retryDelay: number;
    retryStatusCodes: number[];
}

const defaultRetryConfig: RetryConfig = {
    maxRetries: 3,
    retryDelay: 1000,
    retryStatusCodes: [408, 429, 500, 502, 503, 504],
};

const defaultHeaders = {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
};

const defaultOptions = {
    credentials: 'include' as const,
    mode: 'cors' as const
};

interface VerifyTokenResponse {
  valid: boolean;
  message?: string;
}

export class ApiError extends Error {
    status: number;
    code?: string;
    constructor(message: string, status: number, code?: string) {
        super(message);
        this.name = "ApiError";
        this.status = status;
        this.code = code;
    }
}

export class ApiClient {
    private token: string | null = null;
    private tokenExpirationTime: number | null = null;
    private sessionId: string | null = null;
    private refreshPromise: Promise<void> | null = null;

    constructor() {
        if (browser) {
            this.initializeToken();
        }
    }

    public async initializeToken(token?: string, expirationTime?: number, sessionId?: string): Promise<void> {
        try {
            if (token && expirationTime && sessionId) {
                // If token data is provided, use it directly
                await this.saveToken(token, expirationTime, sessionId);
            } else {
                // Otherwise load from storage
                const storedToken = await secureStorage.getSecureItem('token');
                const storedExpiration = await secureStorage.getSecureItem('tokenExpiration');
                const storedSessionId = await secureStorage.getSecureItem('sessionId');
                
                if (storedToken) {
                    this.token = storedToken;
                }
                if (storedExpiration) {
                    this.tokenExpirationTime = parseInt(storedExpiration, 10);
                }
                if (storedSessionId) {
                    this.sessionId = storedSessionId;
                }
            }
        } catch (error) {
            console.error('Failed to initialize token:', error);
        }
    }

    private async saveToken(token: string, expirationTime: number, sessionId: string): Promise<void> {
        try {
            await secureStorage.setSecureItem('token', token);
            await secureStorage.setSecureItem('tokenExpiration', expirationTime.toString());
            await secureStorage.setSecureItem('sessionId', sessionId);
            this.token = token;
            this.tokenExpirationTime = expirationTime;
            this.sessionId = sessionId;
        } catch (error) {
            console.error('Failed to save token:', error);
            throw new ApiError('Failed to save token', 500, ErrorCodes.UNKNOWN_ERROR);
        }
    }

    private async fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
        const headers = new Headers(options.headers);
        options.credentials = 'include';

        // Check if token is close to expiring and refresh if needed
        if (this.token && this.tokenExpirationTime && this.tokenExpirationTime < Date.now() + 15 * 60 * 1000) {
            try {
                console.log('Token close to expiring, attempting refresh...');
                const { token, session } = await this.refreshToken();
                await this.saveToken(token, Date.now() + (24 * 60 * 60 * 1000), session.id); // 24 hours
                this.token = token;
                this.sessionId = session.id;
            } catch (error) {
                console.error('Failed to refresh token:', error);
                // Clear auth state and redirect to login
                await auth.clearAuth();
                throw new ApiError('Session expired', 401, ErrorCodes.SESSION_EXPIRED);
            }
        }

        if (this.token) {
            console.log('Using token for request:', this.token.substring(0, 10) + '...');
            headers.set('Authorization', `Bearer ${this.token}`);
        }

        const csrfToken = getCsrfToken();
        if (csrfToken) {
            console.log('Using CSRF token:', csrfToken);
            headers.set('X-CSRF-Token', csrfToken);
        }

        let response = await fetch(url, { ...options, headers });

        // Handle 401 errors by attempting token refresh
        if (response.status === 401 && this.token) {
            try {
                console.log('Received 401, attempting token refresh...');
                const { token, session } = await this.refreshToken();
                await this.saveToken(token, Date.now() + (24 * 60 * 60 * 1000), session.id); // 24 hours
                headers.set('Authorization', `Bearer ${token}`);
                response = await fetch(url, { ...options, headers });
            } catch (error) {
                console.error('Token refresh failed:', error);
                // Clear auth state and redirect to login
                await auth.clearAuth();
                throw new ApiError('Session expired', 401, ErrorCodes.SESSION_EXPIRED);
            }
        }

        if (response.status === 401) {
            // Clear auth state and redirect to login
            await auth.clearAuth();
            throw new ApiError('Session expired', 401, ErrorCodes.SESSION_EXPIRED);
        }

        return response;
    }

    async verifyToken(token: string): Promise<VerifyTokenResponse> {
        try {
            const response = await fetch(`${API_URL}/api/auth/verify`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                credentials: 'include'
            });

            if (!response.ok) {
                throw new ApiError('Token verification failed', response.status);
            }

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Token verification error:', error);
            throw error;
        }
    }

    async login(credentials: { email: string; password: string; rememberMe?: boolean }): Promise<AuthResponse> {
        const response = await fetch(`${API_URL}/api/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(credentials)
        });

        if (!response.ok) {
            throw new ApiError('Login failed', response.status);
        }

        const data = await response.json();
        
        // Store token and session in secureStorage
        if (data.token && data.session) {
            await this.saveToken(
                data.token,
                Date.now() + (30 * 24 * 60 * 60 * 1000), // 30 days
                data.session.id
            );
        }
        
        return data;
    }

    async register(credentials: { email: string; password: string }): Promise<AuthResponse> {
        const response = await fetch(`${API_URL}/api/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify(credentials),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to register');
        }

        return response.json();
    }

    async getGoogleAuthUrl(): Promise<string> {
        return `${API_URL}/api/auth/google`;
    }

    async getGithubAuthUrl(): Promise<string> {
        return `${API_URL}/api/auth/github`;
    }

    async logout(): Promise<void> {
        const response = await fetch(`${API_URL}/api/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new ApiError('Logout failed', response.status);
        }
    }

    // Sessions
    async getSessions(): Promise<ChatSession[]> {
        const response = await this.fetchWithAuth(`${API_URL}/api/chat/sessions`);
        if (!response.ok) {
            throw new ApiError('Failed to get sessions', response.status);
        }
        return response.json();
    }

    async createSession(data: { title: string; model: string }): Promise<ChatSession> {
        const response = await this.fetchWithAuth(`${API_URL}/api/chat/sessions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            throw new ApiError('Failed to create session', response.status);
        }

        const session = await response.json();
        if (!session || !session.id) {
            throw new ApiError('Invalid session response from server', response.status);
        }

        return session;
    }

    async getSession(id: string): Promise<ChatSession> {
        const response = await this.fetchWithAuth(`${API_URL}/api/chat/sessions/${id}`);
        if (!response.ok) {
            throw new ApiError('Failed to get session', response.status);
        }
        return response.json();
    }

    async deleteSession(id: string): Promise<void> {
        const response = await this.fetchWithAuth(`${API_URL}/api/chat/sessions/${id}`, {
            method: 'DELETE'
        });
        if (!response.ok) {
            throw new ApiError('Failed to delete session', response.status);
        }
    }

    // Messages
    async getSessionMessages(sessionId: string, beforeId?: string): Promise<Message[]> {
        const url = beforeId 
            ? `${API_URL}/api/chat/sessions/${sessionId}/messages?before=${beforeId}`
            : `${API_URL}/api/chat/sessions/${sessionId}/messages`;
        const response = await this.fetchWithAuth(url);
        if (!response.ok) {
            throw new ApiError('Failed to get messages', response.status);
        }
        return response.json();
    }

    async sendMessage(sessionId: string, content: string): Promise<Message> {
        const response = await this.fetchWithAuth(`${API_URL}/api/chat/sessions/${sessionId}/messages`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ content }),
        });

        if (!response.ok) {
            throw new ApiError('Failed to send message', response.status);
        }

        return response.json();
    }

    // Profile Management
    async getProfile(): Promise<User> {
        const response = await fetch(`${API_URL}/api/auth/profile`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to get profile');
        }

        return response.json();
    }

    async updateProfile(data: UpdateProfileRequest): Promise<User> {
        throw new ApiError('Profile update not implemented yet', 501, ErrorCodes.NOT_IMPLEMENTED);
    }

    async updatePreferences(data: UpdatePreferencesRequest): Promise<User> {
        throw new ApiError('Preferences update not implemented yet', 501, ErrorCodes.NOT_IMPLEMENTED);
    }

    async uploadAvatar(formData: FormData): Promise<{ url: string }> {
        throw new ApiError('Avatar upload not implemented yet', 501, ErrorCodes.NOT_IMPLEMENTED);
    }

    // Session Management
    async getUserSessions(): Promise<SessionInfo[]> {
        return [];
    }

    async deleteUserSession(sessionId: string): Promise<void> {
        throw new ApiError('User sessions not implemented yet', 501, ErrorCodes.NOT_IMPLEMENTED);
    }

    async searchMessages(sessionId: string, query: string, limit: number = 50): Promise<Message[]> {
        throw new ApiError('Message search not implemented yet', 501, ErrorCodes.NOT_IMPLEMENTED);
    }

    async refreshToken(): Promise<{ token: string; session: { id: string; expires_at: string } }> {
        try {
            const response = await fetch(`${API_URL}/api/auth/refresh`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({ 
                    token: this.token,
                    session_id: this.sessionId 
                })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new ApiError(error.message || 'Token refresh failed', response.status);
            }

            const data = await response.json();
            if (!data.token || !data.session) {
                throw new ApiError('Invalid refresh response', 500);
            }
            return data;
        } catch (error) {
            console.error('Token refresh error:', error);
            throw error;
        }
    }

    clearToken() {
        this.token = null;
        this.tokenExpirationTime = null;
        this.sessionId = null;
    }
}

export const api = new ApiClient(); 