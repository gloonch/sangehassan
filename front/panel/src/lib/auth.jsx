import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { fetchJSON } from "./api";

const AuthContext = createContext({
  isAuthenticated: false,
  checking: true,
  login: async () => {},
  logout: async () => {}
});

export const AuthProvider = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let mounted = true;
    const checkSession = async () => {
      try {
        await fetchJSON("/api/admin/session");
        if (!mounted) return;
        setIsAuthenticated(true);
      } catch (error) {
        if (!mounted) return;
        setIsAuthenticated(false);
      } finally {
        if (mounted) setChecking(false);
      }
    };
    checkSession();
    return () => {
      mounted = false;
    };
  }, []);

  const value = useMemo(() => {
    const login = async (payload) => {
      await fetchJSON("/api/admin/login", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      setIsAuthenticated(true);
    };

    const logout = async () => {
      await fetchJSON("/api/admin/logout", { method: "POST" });
      setIsAuthenticated(false);
    };

    return { isAuthenticated, checking, login, logout };
  }, [isAuthenticated, checking]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => useContext(AuthContext);
