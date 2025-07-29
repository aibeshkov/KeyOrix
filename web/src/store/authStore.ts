import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { AuthState, User, LoginFormData } from '../types';
import { authService } from '../services/auth';
import {
    persistAuthData,
    clearPersistedAuthData,
    getCurrentAuthState,
    updateTokenExpiry,
    isTokenValid,
    getTimeUntilExpiry
} from '../utils/auth';

interface AuthStore extends AuthState {
    // Actions
    login: (credentials: LoginFormData) => Promise<void>;
    logout: () => Promise<void>;
    refreshToken: () => Promise<void>;
    checkAuth: () => Promise<void>;
    clearError: () => void;
    setUser: (user: User | null) => void;
    setLoading: (loading: boolean) => void;
    setError: (error: string | null) => void;
}

export const useAuthStore = create<AuthStore>()(
    persist(
        (set, get) => ({
            // Initial state
            user: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,

            // Actions
            login: async (credentials: LoginFormData) => {
                set({ isLoading: true, error: null });

                try {
                    const response = await authService.login(credentials);

                    set({
                        user: response.user,
                        token: response.token,
                        isAuthenticated: true,
                        isLoading: false,
                        error: null,
                    });

                    // Persist authentication data
                    persistAuthData({
                        user: response.user,
                        token: response.token,
                        expiresAt: response.expiresAt,
                        rememberMe: credentials.rememberMe,
                    });
                } catch (error) {
                    const errorMessage = error instanceof Error ? error.message : 'Login failed';
                    set({
                        user: null,
                        token: null,
                        isAuthenticated: false,
                        isLoading: false,
                        error: errorMessage,
                    });
                    throw error;
                }
            },

            logout: async () => {
                set({ isLoading: true });

                try {
                    await authService.logout();
                } catch (error) {
                    console.warn('Logout request failed:', error);
                } finally {
                    // Always clear local state regardless of server response
                    set({
                        user: null,
                        token: null,
                        isAuthenticated: false,
                        isLoading: false,
                        error: null,
                    });

                    // Clear stored data
                    clearPersistedAuthData();
                }
            },

            refreshToken: async () => {
                try {
                    const response = await authService.refreshToken();

                    set({
                        token: response.token,
                        error: null,
                    });

                    // Update token expiration
                    updateTokenExpiry(response.expiresAt);
                } catch (error) {
                    // If refresh fails, logout user
                    await get().logout();
                    throw error;
                }
            },

            checkAuth: async () => {
                set({ isLoading: true });

                try {
                    // First check if we have valid persisted data
                    const currentState = getCurrentAuthState();

                    if (currentState.isAuthenticated && isTokenValid()) {
                        // We have valid persisted data, use it
                        set({
                            user: currentState.user,
                            token: currentState.token,
                            isAuthenticated: true,
                            isLoading: false,
                            error: null,
                        });
                        return;
                    }

                    // If no valid persisted data, check with server
                    const user = await authService.checkAuth();

                    if (user) {
                        set({
                            user,
                            isAuthenticated: true,
                            isLoading: false,
                            error: null,
                        });
                    } else {
                        // Clear any invalid persisted data
                        clearPersistedAuthData();
                        set({
                            user: null,
                            token: null,
                            isAuthenticated: false,
                            isLoading: false,
                            error: null,
                        });
                    }
                } catch (error) {
                    // Clear any invalid persisted data
                    clearPersistedAuthData();
                    set({
                        user: null,
                        token: null,
                        isAuthenticated: false,
                        isLoading: false,
                        error: null,
                    });
                }
            },

            clearError: () => {
                set({ error: null });
            },

            setUser: (user: User | null) => {
                set({ user, isAuthenticated: !!user });
            },

            setLoading: (loading: boolean) => {
                set({ isLoading: loading });
            },

            setError: (error: string | null) => {
                set({ error });
            },
        }),
        {
            name: 'auth-storage',
            partialize: (state) => ({
                user: state.user,
                token: state.token,
                isAuthenticated: state.isAuthenticated,
            }),
        }
    )
);

// Helper function to check if token needs refresh
export const shouldRefreshToken = (): boolean => {
    const timeUntilExpiry = getTimeUntilExpiry();
    const fiveMinutes = 5 * 60 * 1000;

    return timeUntilExpiry > 0 && timeUntilExpiry < fiveMinutes;
};

// Helper function to check if token is expired
export const isTokenExpired = (): boolean => {
    return !isTokenValid();
};