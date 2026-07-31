import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { fetchJSON } from "./api";

const AuthContext = createContext({
  isAuthenticated: false,
  checking: true,
  user: null,
  hasPermission: () => false,
  login: async () => {},
  logout: async () => {}
});

export const AuthProvider = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [checking, setChecking] = useState(true);
  const [user, setUser] = useState(null);

  useEffect(() => {
    let mounted = true;
    const checkSession = async () => {
      try {
        const response = await fetchJSON("/api/v1/session");
        if (!mounted) return;
        const session = response?.data || response;
        setIsAuthenticated(Boolean(session?.authenticated && session?.user?.user_type === "INTERNAL"));
        setUser(session?.authenticated ? session.user : null);
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
      const response = await fetchJSON("/api/v1/auth/internal/login", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      const nextUser = response?.data || response;
      if (nextUser?.user_type !== "INTERNAL") throw new Error("internal account required");
      setUser(nextUser);
      setIsAuthenticated(true);
    };

    const logout = async () => {
      await fetchJSON("/api/v1/auth/logout", { method: "POST" });
      setIsAuthenticated(false);
      setUser(null);
    };

    const hasPermission = (code) => Boolean(user?.permissions?.includes(code));
    const refreshUser = async () => {
      const response = await fetchJSON("/api/v1/me");
      const nextUser = response?.data || response;
      setUser(nextUser);
      return nextUser;
    };
    return { isAuthenticated, checking, user, hasPermission, refreshUser, login, logout };
  }, [isAuthenticated, checking, user]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => useContext(AuthContext);
