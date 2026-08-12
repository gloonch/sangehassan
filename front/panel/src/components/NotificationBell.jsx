import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useAuth } from "../lib/auth";

export default function NotificationBell() {
  const { hasPermission } = useAuth();
  const canView = hasPermission("notifications.view_own");
  const [unread, setUnread] = useState(0);
  useEffect(() => {
    if (!canView) return undefined;
    let active = true;
    async function refresh() {
      try {
        const response = await fetchJSON("/api/v1/notifications?limit=1");
        if (active) setUnread(response.data?.unread_count || 0);
      } catch { /* session guard handles auth errors */ }
    }
    refresh();
    const timer = window.setInterval(refresh, 30000);
    return () => { active = false; window.clearInterval(timer); };
  }, [canView]);
  if (!canView) return null;
  return <Link aria-label="اعلان‌ها" to="/dashboard/notifications" className="relative rounded-full border border-primary/20 px-3 py-2 text-lg">🔔{unread > 0 && <span className="absolute -left-2 -top-2 min-w-5 rounded-full bg-red-700 px-1 text-center text-[10px] text-white">{unread}</span>}</Link>;
}
