import { useLocation } from "react-router-dom";
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import AppRoutes from "./router";

export default function App() {
  const location = useLocation();
  const isHome = location.pathname === "/";
  const hideFooter = ["/", "/profile"].includes(location.pathname);

  return (
    <div className="flex min-h-screen flex-col bg-sand text-primary">
      <Navbar />
      <main className={`${isHome ? "pt-0" : "pt-24"} flex-1`}>
        <AppRoutes />
      </main>
      {!hideFooter && <Footer />}
    </div>
  );
}
