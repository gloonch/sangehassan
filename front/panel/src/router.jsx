import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./lib/auth";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Categories from "./pages/Categories";
import Products from "./pages/Products";
import Blocks from "./pages/Blocks";
import Blogs from "./pages/Blogs";
import Projects from "./pages/Projects";
import Templates from "./pages/Templates";
import ContentSections from "./pages/ContentSections";
import Ads from "./pages/Ads";
import Users from "./pages/Users";
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
        <Route path="blocks" element={<Blocks />} />
        <Route path="blogs" element={<Blogs />} />
        <Route path="projects" element={<Projects />} />
        <Route path="templates" element={<Templates />} />
        <Route path="content" element={<ContentSections />} />
        <Route path="ads" element={<Ads />} />
        <Route path="users" element={<Users />} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
