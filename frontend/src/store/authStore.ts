import { create } from 'zustand';
import { Platform } from 'react-native';
import { UserProfile } from '../../proto/auth';
import { BYPASS_AUTH } from '../utils/config';

interface AuthState {
    user: UserProfile | null;
    token: string | null;
    isLoading: boolean;
    login: (user: UserProfile, token: string) => Promise<void>;
    logout: () => Promise<void>;
    restoreSession: () => Promise<void>;
}

// Storage helpers that work on both web and native
const storage = {
    setItem: async (key: string, value: string) => {
        if (Platform.OS === 'web') {
            localStorage.setItem(key, value);
        } else {
            const SecureStore = await import('expo-secure-store');
            await SecureStore.setItemAsync(key, value);
        }
    },
    getItem: async (key: string): Promise<string | null> => {
        if (Platform.OS === 'web') {
            return localStorage.getItem(key);
        } else {
            const SecureStore = await import('expo-secure-store');
            return await SecureStore.getItemAsync(key);
        }
    },
    deleteItem: async (key: string) => {
        if (Platform.OS === 'web') {
            localStorage.removeItem(key);
        } else {
            const SecureStore = await import('expo-secure-store');
            await SecureStore.deleteItemAsync(key);
        }
    }
};

export const useAuthStore = create<AuthState>((set) => ({
    user: null,
    token: null,
    isLoading: true,
    login: async (user, token) => {
        await storage.setItem('auth_token', token);
        await storage.setItem('user_profile', JSON.stringify(user));
        set({ user, token, isLoading: false });
    },
    logout: async () => {
        await storage.deleteItem('auth_token');
        await storage.deleteItem('user_profile');
        set({ user: null, token: null, isLoading: false });
    },
    restoreSession: async () => {
        console.log('[AUTH] Starting session restore...');
        console.log('[AUTH] BYPASS_AUTH =', BYPASS_AUTH);
        try {
            const token = await storage.getItem('auth_token');
            const userStr = await storage.getItem('user_profile');
            console.log('[AUTH] Retrieved from storage - token:', !!token, 'user:', !!userStr);

            if (token && userStr) {
                const user = JSON.parse(userStr) as UserProfile;
                console.log('[AUTH] Restoring existing session for user:', user.email);
                set({ user, token, isLoading: false });
            } else if (BYPASS_AUTH) {
                // Bypass authentication for testing
                console.log('[BYPASS] No session found, auto-login with mock user');
                const mockUser = {
                    id: 'mock-user-id',
                    email: 'test@example.com',
                    name: 'Test User',
                    picture: '',
                };
                const mockToken = 'mock-jwt-token';
                await storage.setItem('auth_token', mockToken);
                await storage.setItem('user_profile', JSON.stringify(mockUser));
                console.log('[BYPASS] Mock user saved to storage, setting state...');
                set({ user: mockUser, token: mockToken, isLoading: false });
                console.log('[BYPASS] State updated, isLoading should be false now');
            } else {
                console.log('[AUTH] No session found, no bypass, showing login');
                set({ isLoading: false });
            }
        } catch (e) {
            console.error("[AUTH] Failed to restore session", e);
            set({ isLoading: false });
        }
    },
}));
