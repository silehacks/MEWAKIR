import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { AppProvider, useAppContext } from './context/AppContext';
import { LoginPage } from './pages/LoginPage';
import { HomePage } from './pages/HomePage';
import { MeetingPage } from './pages/MeetingPage';
import './styles/global.css';

function ProtectedRoute({ children }) {
  const { displayName } = useAppContext();
  if (!displayName) {
    return <Navigate to="/" replace />;
  }
  return children;
}

export function App() {
  return (
    <AppProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<LoginPage />} />
          <Route
            path="/home"
            element={
              <ProtectedRoute>
                <HomePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/meeting/:meetingId"
            element={
              <ProtectedRoute>
                <MeetingPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </BrowserRouter>
    </AppProvider>
  );
}
