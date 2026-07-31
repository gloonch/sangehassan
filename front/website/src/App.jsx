import { useLocation } from "react-router-dom";
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import GoogleAnalytics from "./components/GoogleAnalytics";
import AppRoutes from "./router";

export default function App() {
  const location = useLocation();
  const isHome = location.pathname === "/";
  const hideFooter = ["/profile", "/login", "/signup"].includes(location.pathname);
  const mainTopClass = isHome ? "pt-0" : "pt-28";

  return (
    <div className="flex min-h-screen flex-col bg-sand text-primary">
      <GoogleAnalytics />
      <Navbar />
      <main className={`${mainTopClass} flex-1`}>
        <AppRoutes />
      </main>
      {!hideFooter && <Footer />}
    </div>
  );
}
