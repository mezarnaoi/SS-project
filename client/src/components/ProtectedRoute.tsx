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
    return <div className="flex justify-center items-center min-h-[60vh]">
      <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-sky-500"></div>
    </div>;
  }

  if (authRequired && !isLoggedIn) {
    return <Navigate to="/login" replace />;
  }

  if (!authRequired && isLoggedIn) {
    return <Navigate to="/" replace />;
  }

  if (authRequired && requiredPage && !hasPageAccess(requiredPage)) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
};

export default ProtectedRoute; 
