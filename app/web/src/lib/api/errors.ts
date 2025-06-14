export class ApiError extends Error {
    public status?: number;
    public code?: string;
    public details?: any;

    constructor(message: string, status?: number, code?: string, details?: any) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.code = code;
        this.details = details;
    }
}

export class NetworkError extends Error {
    constructor(message: string = 'Network error occurred') {
        super(message);
        this.name = 'NetworkError';
    }
}

export const ErrorCodes = {
    SESSION_EXPIRED: 'SESSION_EXPIRED',
    NO_TOKEN: 'NO_TOKEN',
    REFRESH_FAILED: 'REFRESH_FAILED',
    UNKNOWN_ERROR: 'UNKNOWN_ERROR',
    VALIDATION_ERROR: 'VALIDATION_ERROR',
    RATE_LIMIT: 'RATE_LIMIT',
    SERVER_ERROR: 'SERVER_ERROR',
    CSRF_ERROR: 'CSRF_ERROR',
    INVALID_TOKEN: 'INVALID_TOKEN',
    NOT_IMPLEMENTED: 'NOT_IMPLEMENTED',
} as const;

export type ErrorCode = typeof ErrorCodes[keyof typeof ErrorCodes]; 