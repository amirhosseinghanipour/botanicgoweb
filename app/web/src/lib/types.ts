export interface User {
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

export interface Message {
    id: string;
    session_id: string;
    user_id: string;
    content: string;
    model: string;
    type: 'message' | 'error' | 'typing' | 'status';
    created_at: string;
    updated_at: string;
    role: 'user' | 'assistant';
}

export interface ChatSession {
    id: string;
    user_id: string;
    title: string;
    model: string;
    created_at: string;
    updated_at: string;
    messages: Message[];
} 