import { useLocation } from "react-router-dom";
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import AppRoutes from "./router";

export default function App() {
  const location = useLocation();
  const hideFooter = location.pathname === "/";

  return (
    <div className="min-h-screen bg-sand text-primary">
      <Navbar />
      <main className="pt-24">
        <AppRoutes />
      </main>
      {!hideFooter && <Footer />}
    </div>
  );
}
