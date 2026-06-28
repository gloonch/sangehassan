import { Navigate, Route, Routes } from "react-router-dom";
import Home from "./pages/Home";
import ProductsLanding from "./pages/ProductsLanding";
import ProductDetail from "./pages/ProductDetail";
import Blogs from "./pages/Blogs";
import BlogDetail from "./pages/BlogDetail";
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
import ProductHub from "./pages/ProductHub";
import ProductCatalog from "./pages/ProductCatalog";
import LocalizedProductRoute from "./pages/LocalizedProductRoute";
import NotFound from "./pages/NotFound";

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/products" element={<ProductsLanding />} />
      <Route path="/products/overview" element={<Navigate to="/products" replace />} />
      <Route path="/products/:slug" element={<ProductDetail />} />
      <Route path="/fa/products" element={<ProductHub />} />
      <Route path="/fa/products/:slug" element={<LocalizedProductRoute />} />
      <Route path="/fa/products/:categorySlug/:facet/:value" element={<ProductCatalog />} />
      <Route path="/en/products" element={<ProductHub />} />
      <Route path="/en/products/:slug" element={<LocalizedProductRoute />} />
      <Route path="/en/products/:categorySlug/:facet/:value" element={<ProductCatalog />} />
      <Route path="/ar/products" element={<ProductHub />} />
      <Route path="/ar/products/:slug" element={<LocalizedProductRoute />} />
      <Route path="/ar/products/:categorySlug/:facet/:value" element={<ProductCatalog />} />
      <Route path="/blocks" element={<BlocksLanding />} />
      <Route path="/blocks/:slug" element={<BlockDetail />} />
      <Route path="/ads" element={<Ads />} />
      <Route path="/ads/:id" element={<AdDetail />} />
      <Route element={<RequireUserAuth />}>
        <Route path="/ads/new" element={<NewAd />} />
      </Route>
      <Route path="/blogs" element={<Blogs />} />
      <Route path="/fa/blogs" element={<Blogs />} />
      <Route path="/fa/blogs/:slug" element={<BlogDetail />} />
      <Route path="/en/blogs" element={<Blogs />} />
      <Route path="/en/blogs/:slug" element={<BlogDetail />} />
      <Route path="/ar/blogs" element={<Blogs />} />
      <Route path="/ar/blogs/:slug" element={<BlogDetail />} />
      <Route path="/projects" element={<Projects />} />
      <Route path="/projects/:id" element={<ProjectDetail />} />
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Navigate to="/login?mode=signup" replace />} />
      <Route path="/profile" element={<Profile />} />
      <Route path="/gallery" element={<Gallery />} />
      <Route path="/about" element={<About />} />
      <Route path="/404" element={<NotFound />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}
