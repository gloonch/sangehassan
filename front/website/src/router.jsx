import { Navigate, Route, Routes } from "react-router-dom";
import Home from "./pages/Home";
import ProductsLanding from "./pages/ProductsLanding";
import ProductDetail from "./pages/ProductDetail";
import Blogs from "./pages/Blogs";
import Login from "./pages/Login";
import Profile from "./pages/Profile";
import Gallery from "./pages/Gallery";
import About from "./pages/About";
import BlocksLanding from "./pages/BlocksLanding";
import BlockDetail from "./pages/BlockDetail";
import Ads from "./pages/Ads";
import AdDetail from "./pages/AdDetail";
import NewAd from "./pages/NewAd";
import Projects from "./pages/Projects";
import ProjectDetail from "./pages/ProjectDetail";
import RequireUserAuth from "./components/RequireUserAuth";

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/products" element={<ProductsLanding />} />
      <Route path="/products/overview" element={<Navigate to="/products" replace />} />
      <Route path="/products/:slug" element={<ProductDetail />} />
      <Route path="/blocks" element={<BlocksLanding />} />
      <Route path="/blocks/:slug" element={<BlockDetail />} />
      <Route path="/ads" element={<Ads />} />
      <Route path="/ads/:id" element={<AdDetail />} />
      <Route element={<RequireUserAuth />}>
        <Route path="/ads/new" element={<NewAd />} />
      </Route>
      <Route path="/blogs" element={<Blogs />} />
      <Route path="/projects" element={<Projects />} />
      <Route path="/projects/:id" element={<ProjectDetail />} />
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Navigate to="/login?mode=signup" replace />} />
      <Route path="/profile" element={<Profile />} />
      <Route path="/gallery" element={<Gallery />} />
      <Route path="/about" element={<About />} />
    </Routes>
  );
}
