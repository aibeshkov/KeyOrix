import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { useAuthStore, shouldRefreshToken, isTokenExpired } from '../store/authStore';
import { getEnvConfig } from '../utils';

const config = getEnvConfig();

// Create main API client
export const apiClient: AxiosInstance = axios.create({
    baseURL: config.API_BASE_URL,
    timeout: config.API_TIMEOUT,
    withCredentials: true, // Important for HTTP-only cookies
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor for authentication
apiClient.interceptors.request.use(
    async (config) => {
        const authStore = useAuthStore.getState();

        // Skip auth for login/refresh endpoints
        const isAuthEndpoint = config.url?.includes('/auth/login') ||
            config.url?.includes('/auth/refresh');

        if (!isAuthEndpoint && authStore.isAuthenticated) {
            // Check if token is expired
            if (isTokenExpired()) {
                // Token is expired, logout user
                await authStore.logout();
                throw new Error('Session expired');
            }

            // Check if token needs refresh
            if (shouldRefreshToken()) {
                try {
                    await authStore.refreshToken();
                } catch (error) {
                    // Refresh failed, logout will be handled by the store
                    throw error;
                }
            }

            // Add authorization header if we have a token
            if (authStore.token) {
                config.headers.Authorization = `Bearer ${authStore.token}`;
            }
        }

        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor for error handling
apiClient.interceptors.response.use(
    (response: AxiosResponse) => {
        return response;
    },
    async (error: AxiosError) => {
        const authStore = useAuthStore.getState();

        if (error.response?.status === 401) {
            // Unauthorized - token might be invalid or expired
            if (authStore.isAuthenticated) {
                // Try to refresh token once
                if (!error.config?.url?.includes('/auth/refresh')) {
                    try {
                        await authStore.refreshToken();

                        // Retry the original request with new token
                        if (error.config) {
                            const newToken = useAuthStore.getState().token;
                            if (newToken) {
                                error.config.headers.Authorization = `Bearer ${newToken}`;
                                return apiClient.request(error.config);
                            }
                        }
                    } catch (refreshError) {
                        // Refresh failed, logout user
                        await authStore.logout();
                        authStore.setError('Your session has expired. Please log in again.');
                    }
                } else {
                    // Refresh endpoint failed, logout user
                    await authStore.logout();
                    authStore.setError('Your session has expired. Please log in again.');
                }
            }
        } else if (error.response?.status === 403) {
            // Forbidden - user doesn't have permission
            authStore.setError('You do not have permission to perform this action.');
        } else if (error.response?.status >= 500) {
            // Server error
            authStore.setError('Server error. Please try again later.');
        } else if (error.code === 'ECONNABORTED') {
            // Request timeout
            authStore.setError('Request timeout. Please check your connection and try again.');
        } else if (!error.response) {
            // Network error
            authStore.setError('Network error. Please check your connection.');
        }

        return Promise.reject(error);
    }
);

// Helper function to make authenticated requests
export const makeAuthenticatedRequest = async <T>(
    config: AxiosRequestConfig
): Promise<T> => {
    try {
        const response = await apiClient.request<T>(config);
        return response.data;
    } catch (error) {
        throw error;
    }
};

// Helper function to handle API errors consistently
export const handleApiError = (error: unknown): string => {
    if (axios.isAxiosError(error)) {
        if (error.response?.data?.error) {
            return error.response.data.error;
        }
        if (error.response?.data?.message) {
            return error.response.data.message;
        }
        if (error.message) {
            return error.message;
        }
    }

    if (error instanceof Error) {
        return error.message;
    }

    return 'An unexpected error occurred';
};

// Export configured axios instance
export default apiClient;