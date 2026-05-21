import { BrowserRouter, Routes, Route, Navigate, Outlet, useNavigate } from 'react-router-dom';
import Navbar from './components/navbar';
import HomePage from './pages/homePage';
import LoginPage from './pages/loginPage';
import RegisterPage from './pages/register';
import PhotosPage from './pages/photosPage';
import DevicesPage from './pages/devicesPage';
import StatisticsPage from './pages/statisticsPage';
import ReportsPage from './pages/reportsPage';
import ProtectedRoute from './components/ProtectedRoute';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ThemeProvider, useTheme } from './contexts/ThemeContext';

const Layout = () => {
  const navigate = useNavigate();
  const { isLoggedIn, logout } = useAuth();
  const { isDark, toggleTheme } = useTheme();

  // Left-side buttons (only shown when logged in)
  const leftButtons = isLoggedIn
    ? [
      {
        text: 'Photos',
        variant: 'secondary' as const,
        onClick: () => navigate('/photos')
      },
      {
        text: 'Devices',
        variant: 'secondary' as const,
        onClick: () => navigate('/devices')
      },
      {
        text: 'Statistics',
        variant: 'secondary' as const,
        onClick: () => navigate('/statistics')
      },
      {
        text: 'Reports',
        variant: 'secondary' as const,
        onClick: () => navigate('/reports')
      }
    ]
    : [];

  // Right-side buttons (different based on login status)
  const rightButtons = isLoggedIn
    ? [
      {
        text: isDark ? '☀️ Light' : '🌙 Dark',
        variant: 'outline' as const,
        onClick: toggleTheme
        },
      {
        text: 'Logout',
        variant: 'outline' as const,
        onClick: () => {
          logout();
          navigate('/');
        }
      }
    ]
    : [
      {
        text: isDark ? '☀️ Light' : '🌙 Dark',
        variant: 'outline' as const,
        onClick: toggleTheme
      },
      {
        text: 'Login',
        variant: 'outline' as const,
        onClick: () => navigate('/login')
      },
      {
        text: 'Register',
        variant: 'primary' as const,
        onClick: () => navigate('/register')
      }
    ];

  return (
    <>
      <Navbar
        title="Security of Systems - First Force"
        leftButtons={leftButtons}
        rightButtons={rightButtons}
      />
      <div className="pt-16 px-4 min-h-screen bg-white dark:bg-gray-900 dark:text-gray-100 transition-colors">
        <Outlet />
      </div>
    </>
  );
};

const App = () => {
  return (
    <BrowserRouter>
      <ThemeProvider>  
        <AuthProvider>
          <Routes>
          {/* Common layout for all routes */}
          <Route element={<Layout />}>
            {/* Public routes - accessible to everyone */}
            <Route path="/" element={<HomePage />} />

            {/* Auth routes - only for non-authenticated users */}
            <Route element={<ProtectedRoute authRequired={false} />}>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
            </Route>

            {/* Protected routes - only for authenticated users */}
            <Route element={<ProtectedRoute authRequired={true} />}>
              <Route path="/photos" element={<PhotosPage />} />
              <Route path="/devices" element={<DevicesPage />} />
              <Route path="/statistics" element={<StatisticsPage />} />
              <Route path="/reports" element={<ReportsPage />} />
            </Route>

            {/* Fallback route */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
          </Routes>
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
};

export default App;
