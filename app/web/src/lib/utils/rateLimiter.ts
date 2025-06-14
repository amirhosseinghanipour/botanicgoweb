interface RateLimitConfig {
    maxRequests: number;
    timeWindow: number; // in milliseconds
}

interface RateLimitState {
    requests: number;
    resetTime: number;
}

class RateLimiter {
    private state: Map<string, RateLimitState> = new Map();
    private config: RateLimitConfig;

    constructor(config: RateLimitConfig) {
        this.config = config;
    }

    private getKey(endpoint: string): string {
        return endpoint;
    }

    private cleanup(): void {
        const now = Date.now();
        for (const [key, state] of this.state.entries()) {
            if (now > state.resetTime) {
                this.state.delete(key);
            }
        }
    }

    canMakeRequest(endpoint: string): boolean {
        this.cleanup();
        const key = this.getKey(endpoint);
        const now = Date.now();
        const state = this.state.get(key);

        if (!state) {
            this.state.set(key, {
                requests: 1,
                resetTime: now + this.config.timeWindow
            });
            return true;
        }

        if (now > state.resetTime) {
            this.state.set(key, {
                requests: 1,
                resetTime: now + this.config.timeWindow
            });
            return true;
        }

        if (state.requests >= this.config.maxRequests) {
            return false;
        }

        state.requests++;
        return true;
    }

    getRemainingRequests(endpoint: string): number {
        const key = this.getKey(endpoint);
        const state = this.state.get(key);
        if (!state) return this.config.maxRequests;
        return Math.max(0, this.config.maxRequests - state.requests);
    }

    getResetTime(endpoint: string): number {
        const key = this.getKey(endpoint);
        const state = this.state.get(key);
        if (!state) return 0;
        return state.resetTime;
    }
}

// Create a singleton instance with default configuration
export const rateLimiter = new RateLimiter({
    maxRequests: 100, // 100 requests
    timeWindow: 60000 // per minute
}); 