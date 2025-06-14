import { browser } from '$app/environment';

const COOKIE_OPTIONS = {
    path: '/',
    secure: true,
    sameSite: 'strict' as const,
    maxAge: 7 * 24 * 60 * 60 // 7 days
};

export class SecureStorage {
    private static instance: SecureStorage;
    private encryptionKey: CryptoKey | null = null;
    private initializationPromise: Promise<void> | null = null;

    private constructor() {
        if (browser) {
            this.initializationPromise = this.initializeEncryption();
        }
    }

    static getInstance(): SecureStorage {
        if (!SecureStorage.instance) {
            SecureStorage.instance = new SecureStorage();
        }
        return SecureStorage.instance;
    }

    private async initializeEncryption(): Promise<void> {
        try {
            // Generate a key from a password
            const password = await this.getOrCreatePassword();
            const encoder = new TextEncoder();
            const keyMaterial = await crypto.subtle.importKey(
                'raw',
                encoder.encode(password),
                { name: 'PBKDF2' },
                false,
                ['deriveBits', 'deriveKey']
            );

            this.encryptionKey = await crypto.subtle.deriveKey(
                {
                    name: 'PBKDF2',
                    salt: encoder.encode('botanic-salt'),
                    iterations: 100000,
                    hash: 'SHA-256'
                },
                keyMaterial,
                { name: 'AES-GCM', length: 256 },
                false,
                ['encrypt', 'decrypt']
            );
        } catch (error) {
            console.error('Failed to initialize encryption:', error);
            throw error;
        }
    }

    private async getOrCreatePassword(): Promise<string> {
        let password = localStorage.getItem('encryption_password');
        if (!password) {
            password = crypto.randomUUID();
            localStorage.setItem('encryption_password', password);
        }
        return password;
    }

    async setSecureItem(key: string, value: string): Promise<void> {
        if (!browser) return;

        try {
            // Wait for initialization to complete
            if (this.initializationPromise) {
                await this.initializationPromise;
            }

            if (!this.encryptionKey) {
                throw new Error('Encryption not initialized');
            }

            const encoder = new TextEncoder();
            const iv = crypto.getRandomValues(new Uint8Array(12));
            const encryptedData = await crypto.subtle.encrypt(
                {
                    name: 'AES-GCM',
                    iv
                },
                this.encryptionKey,
                encoder.encode(value)
            );

            const encryptedValue = btoa(
                String.fromCharCode(...new Uint8Array(encryptedData))
            );
            const ivString = btoa(String.fromCharCode(...iv));

            localStorage.setItem(key, JSON.stringify({
                data: encryptedValue,
                iv: ivString
            }));
        } catch (error) {
            console.error('Failed to encrypt data:', error);
            throw new Error('Failed to encrypt data');
        }
    }

    async getSecureItem(key: string): Promise<string | null> {
        if (!browser) return null;

        try {
            // Wait for initialization to complete
            if (this.initializationPromise) {
                await this.initializationPromise;
            }

            if (!this.encryptionKey) {
                throw new Error('Encryption not initialized');
            }

            const item = localStorage.getItem(key);
            if (!item) return null;

            const { data, iv } = JSON.parse(item);
            const encryptedData = Uint8Array.from(atob(data), c => c.charCodeAt(0));
            const ivData = Uint8Array.from(atob(iv), c => c.charCodeAt(0));

            const decryptedData = await crypto.subtle.decrypt(
                {
                    name: 'AES-GCM',
                    iv: ivData
                },
                this.encryptionKey,
                encryptedData
            );

            return new TextDecoder().decode(decryptedData);
        } catch (error) {
            console.error('Failed to decrypt data:', error);
            return null;
        }
    }

    removeSecureItem(key: string): void {
        if (!browser) return;
        localStorage.removeItem(key);
    }

    setSecureCookie(name: string, value: string, options = COOKIE_OPTIONS): void {
        if (!browser) return;

        const cookieValue = encodeURIComponent(value);
        const cookieOptions = Object.entries(options)
            .map(([key, value]) => `${key}=${value}`)
            .join('; ');

        document.cookie = `${name}=${cookieValue}; ${cookieOptions}`;
    }

    getSecureCookie(name: string): string | null {
        if (!browser) return null;

        const cookies = document.cookie.split(';');
        const cookie = cookies.find(c => c.trim().startsWith(`${name}=`));
        if (!cookie) return null;

        return decodeURIComponent(cookie.split('=')[1]);
    }

    removeSecureCookie(name: string): void {
        if (!browser) return;
        this.setSecureCookie(name, '', { ...COOKIE_OPTIONS, maxAge: 0 });
    }
}

export const secureStorage = SecureStorage.getInstance(); 