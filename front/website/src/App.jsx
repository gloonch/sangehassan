import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import AppRoutes from "./router";

export default function App() {
  return (
    <div className="min-h-screen bg-sand text-primary">
      <Navbar />
      <main className="pt-24">
        <AppRoutes />
      </main>
      <Footer />
    </div>
  );
}
