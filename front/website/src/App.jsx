import { useLocation } from "react-router-dom";
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import RouteRevealOverlay from "./components/RouteRevealOverlay";
import AppRoutes from "./router";

export default function App() {
  const location = useLocation();
  const isHome = location.pathname === "/";
  const isAuthPage = ["/login", "/signup"].includes(location.pathname);
  const hideFooter = ["/", "/profile", "/login", "/signup"].includes(location.pathname);

  return (
    <div className="flex min-h-screen flex-col bg-sand text-primary">
      <RouteRevealOverlay />
      <Navbar />
      <main className={`${isHome ? "pt-0" : isAuthPage ? "pt-20" : "pt-24"} flex-1`}>
        <AppRoutes />
      </main>
      {!hideFooter && <Footer />}
    </div>
  );
}
