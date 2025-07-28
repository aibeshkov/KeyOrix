import React from 'react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../../hooks/useAuth';

export const DashboardPage: React.FC = () => {
    const { t } = useTranslation();
    const { user, logout } = useAuth();

    const handleLogout = async () => {
        await logout();
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <div className="container mx-auto px-4 py-8">
                <header className="mb-8">
                    <div className="flex justify-between items-center">
                        <div>
                            <h1 className="text-3xl font-bold text-gray-900">
                                {t('dashboard.title')}
                            </h1>
                            <p className="text-gray-600 mt-2">
                                {t('dashboard.welcome')}, {user?.username}!
                            </p>
                        </div>
                        <button
                            onClick={handleLogout}
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        >
                            {t('auth.logout')}
                        </button>
                    </div>
                </header>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {/* Stats Cards */}
                    <div className="bg-white rounded-lg shadow p-6">
                        <h3 className="text-lg font-medium text-gray-900 mb-2">
                            {t('dashboard.totalSecrets')}
                        </h3>
                        <p className="text-3xl font-bold text-blue-600">0</p>
                        <p className="text-sm text-gray-500 mt-1">
                            No secrets yet
                        </p>
                    </div>

                    <div className="bg-white rounded-lg shadow p-6">
                        <h3 className="text-lg font-medium text-gray-900 mb-2">
                            {t('dashboard.sharedSecrets')}
                        </h3>
                        <p className="text-3xl font-bold text-green-600">0</p>
                        <p className="text-sm text-gray-500 mt-1">
                            No shared secrets
                        </p>
                    </div>

                    <div className="bg-white rounded-lg shadow p-6">
                        <h3 className="text-lg font-medium text-gray-900 mb-2">
                            {t('dashboard.recentActivity')}
                        </h3>
                        <p className="text-3xl font-bold text-purple-600">0</p>
                        <p className="text-sm text-gray-500 mt-1">
                            No recent activity
                        </p>
                    </div>
                </div>

                {/* Quick Actions */}
                <div className="mt-8">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">
                        {t('dashboard.quickActions')}
                    </h2>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                        <button className="p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow text-left">
                            <h3 className="font-medium text-gray-900">
                                {t('dashboard.createSecret')}
                            </h3>
                            <p className="text-sm text-gray-500 mt-1">
                                Create a new secret
                            </p>
                        </button>

                        <button className="p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow text-left">
                            <h3 className="font-medium text-gray-900">
                                {t('dashboard.viewSecrets')}
                            </h3>
                            <p className="text-sm text-gray-500 mt-1">
                                View all your secrets
                            </p>
                        </button>

                        <button className="p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow text-left">
                            <h3 className="font-medium text-gray-900">
                                {t('sharing.title')}
                            </h3>
                            <p className="text-sm text-gray-500 mt-1">
                                Manage shared secrets
                            </p>
                        </button>

                        <button className="p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow text-left">
                            <h3 className="font-medium text-gray-900">
                                {t('profile.title')}
                            </h3>
                            <p className="text-sm text-gray-500 mt-1">
                                Update your profile
                            </p>
                        </button>
                    </div>
                </div>

                {/* User Info */}
                <div className="mt-8 bg-white rounded-lg shadow p-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">
                        User Information
                    </h2>
                    <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div>
                            <dt className="text-sm font-medium text-gray-500">Username</dt>
                            <dd className="text-sm text-gray-900">{user?.username}</dd>
                        </div>
                        <div>
                            <dt className="text-sm font-medium text-gray-500">Email</dt>
                            <dd className="text-sm text-gray-900">{user?.email}</dd>
                        </div>
                        <div>
                            <dt className="text-sm font-medium text-gray-500">Role</dt>
                            <dd className="text-sm text-gray-900">{user?.role}</dd>
                        </div>
                        <div>
                            <dt className="text-sm font-medium text-gray-500">Last Login</dt>
                            <dd className="text-sm text-gray-900">{user?.lastLogin}</dd>
                        </div>
                    </dl>
                </div>
            </div>
        </div>
    );
};