import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./lib/auth";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Categories from "./pages/Categories";
import Products from "./pages/Products";
import Blogs from "./pages/Blogs";
import PanelLayout from "./components/PanelLayout";
import { useTranslation } from "./lib/i18n";

const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, checking } = useAuth();
  const { t } = useTranslation();

  if (checking) {
    return (
      <div className="panel-shell py-16 text-sm text-primary/70">{t("messages.loading")}</div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <PanelLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="categories" element={<Categories />} />
        <Route path="products" element={<Products />} />
        <Route path="blogs" element={<Blogs />} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
