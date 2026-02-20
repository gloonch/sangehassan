import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

export default function RequireUserAuth() {
  const { t } = useTranslation();
  const location = useLocation();
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    let active = true;
    const check = async () => {
      try {
        await fetchJSON("/api/v1/me");
        if (!active) return;
        setIsAuthenticated(true);
      } catch (_) {
        if (!active) return;
        setIsAuthenticated(false);
      } finally {
        if (active) setIsChecking(false);
      }
    };
    check();
    return () => {
      active = false;
    };
  }, []);

  if (isChecking) {
    return (
      <section className="section-shell py-16">
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      </section>
    );
  }

  if (!isAuthenticated) {
    const redirectTo = `${location.pathname}${location.search}${location.hash}`;
    sessionStorage.setItem("sh_after_login", redirectTo);
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
