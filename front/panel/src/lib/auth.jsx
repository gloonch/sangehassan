import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { fetchJSON } from "./api";

const AuthContext = createContext({
  isAuthenticated: false,
  checking: true,
  user: null,
  hasPermission: () => false,
	featureEnabled: () => true,
  login: async () => {},
  logout: async () => {}
});

export const AuthProvider = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [checking, setChecking] = useState(true);
  const [user, setUser] = useState(null);
	const [features, setFeatures] = useState({});

  useEffect(() => {
    let mounted = true;
    const checkSession = async () => {
      try {
        const response = await fetchJSON("/api/v1/session");
        if (!mounted) return;
        const session = response?.data || response;
        setIsAuthenticated(Boolean(session?.authenticated && session?.user?.user_type === "INTERNAL"));
        setUser(session?.authenticated ? session.user : null);
		if (session?.authenticated && session?.user?.user_type === "INTERNAL") {
		  try { const flags = await fetchJSON("/api/v1/operations/features"); if (mounted) setFeatures(flags.data || {}); } catch { if (mounted) setFeatures({}); }
		}
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
	  try { const flags = await fetchJSON("/api/v1/operations/features"); setFeatures(flags.data || {}); } catch { setFeatures({}); }
    };

    const logout = async () => {
      await fetchJSON("/api/v1/auth/logout", { method: "POST" });
      setIsAuthenticated(false);
      setUser(null);
	  setFeatures({});
    };

    const hasPermission = (code) => Boolean(user?.permissions?.includes(code));
	const featureEnabled = (key) => features[key] !== false;
    const refreshUser = async () => {
      const response = await fetchJSON("/api/v1/me");
      const nextUser = response?.data || response;
      setUser(nextUser);
      return nextUser;
    };
	return { isAuthenticated, checking, user, hasPermission, featureEnabled, refreshUser, login, logout };
	}, [isAuthenticated, checking, user, features]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => useContext(AuthContext);
