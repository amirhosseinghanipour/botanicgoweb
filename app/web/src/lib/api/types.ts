export interface User {
    id: string;
    email: string;
    name?: string;
    avatar?: string;
}

export interface PageData {
    token: string;
    user: User;
    expiresIn: number;
} 