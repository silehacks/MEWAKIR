import { createContext, useContext, useMemo, useState } from 'react';

const AppContext = createContext(null);

export function AppProvider({ children }) {
  const [displayName, setDisplayName] = useState(() => localStorage.getItem('mewakir:name') || '');
  const [activeMeeting, setActiveMeeting] = useState(null);

  const value = useMemo(
    () => ({
      displayName,
      setDisplayName: (name) => {
        setDisplayName(name);
        if (name) {
          localStorage.setItem('mewakir:name', name);
        } else {
          localStorage.removeItem('mewakir:name');
        }
      },
      activeMeeting,
      setActiveMeeting,
    }),
    [displayName, activeMeeting]
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useAppContext() {
  const ctx = useContext(AppContext);
  if (!ctx) {
    throw new Error('useAppContext must be used within AppProvider');
  }
  return ctx;
}
