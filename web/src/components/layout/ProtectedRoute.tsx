import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { ROUTES } from '../../constants';

interface ProtectedRouteProps {
    children: React.ReactNode;
    requiredPermissions?: string[];
    requiredRole?: string;
    fallbackPath?: string;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
    children,
    requiredPermissions = [],
    requiredRole,
    fallbackPath = ROUTES.LOGIN,
}) => {
    const { isAuthenticated, isLoading, user, hasPermission } = useAuth();
    const location = useLocation();

    // Show loading while checking authentication
    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-600">Loading...</p>
                </div>
            </div>
        );
    }

    // Redirect to login if not authenticated
    if (!isAuthenticated) {
        return (
            <Navigate
                to={fallbackPath}
                state={{ from: location }}
                replace
            />
        );
    }

    // Check role requirement
    if (requiredRole && user?.role !== requiredRole) {
        return (
            <Navigate
                to={ROUTES.DASHBOARD}
                replace
            />
        );
    }

    // Check permission requirements
    if (requiredPermissions.length > 0) {
        const hasAllPermissions = requiredPermissions.every(permission =>
            hasPermission(permission)
        );

        if (!hasAllPermissions) {
            return (
                <Navigate
                    to={ROUTES.DASHBOARD}
                    replace
                />
            );
        }
    }

    return <>{children}</>;
};