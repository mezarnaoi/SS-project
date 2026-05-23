import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

interface ProtectedRouteProps {
  authRequired: boolean;
  requiredPage?: string;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ authRequired, requiredPage }) => {
  const { isLoggedIn, loading, hasPageAccess } = useAuth();

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[60vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-sky-500"></div>
      </div>
    );
  }

  if (authRequired && !isLoggedIn) {
    return <Navigate to="/login" replace />;
  }

  if (!authRequired && isLoggedIn) {
    return <Navigate to="/" replace />;
  }

  if (authRequired && requiredPage && !hasPageAccess(requiredPage)) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="max-w-md rounded-lg border border-amber-300 bg-amber-50 p-4 text-amber-900">
          <h2 className="text-lg font-semibold mb-1">Unauthorized</h2>
          <p className="text-sm">You do not have permission to access this page.</p>
        </div>
      </div>
    );
  }

  return <Outlet />;
};

export default ProtectedRoute; 
