import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import Home from "./pages/Home";
import RequireUserAuth from "./components/RequireUserAuth";

const ProductsLanding = lazy(() => import("./pages/ProductsLanding"));
const ProductDetail = lazy(() => import("./pages/ProductDetail"));
const Blogs = lazy(() => import("./pages/Blogs"));
const BlogDetail = lazy(() => import("./pages/BlogDetail"));
const Login = lazy(() => import("./pages/Login"));
const Profile = lazy(() => import("./pages/Profile"));
const Gallery = lazy(() => import("./pages/Gallery"));
const About = lazy(() => import("./pages/About"));
const BlocksLanding = lazy(() => import("./pages/BlocksLanding"));
const BlockDetail = lazy(() => import("./pages/BlockDetail"));
const Ads = lazy(() => import("./pages/Ads"));
const AdDetail = lazy(() => import("./pages/AdDetail"));
const NewAd = lazy(() => import("./pages/NewAd"));
const Projects = lazy(() => import("./pages/Projects"));
const ProjectDetail = lazy(() => import("./pages/ProjectDetail"));
const ProductHub = lazy(() => import("./pages/ProductHub"));
const ProductCatalog = lazy(() => import("./pages/ProductCatalog"));
const LocalizedProductRoute = lazy(() => import("./pages/LocalizedProductRoute"));
const StoneSampleRequest = lazy(() => import("./pages/StoneSampleRequest"));
const NotFound = lazy(() => import("./pages/NotFound"));

export default function AppRoutes() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-sand" aria-busy="true" />}>
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
        <Route path="/stone-sample-request" element={<StoneSampleRequest />} />
      </Route>
      <Route path="/blogs" element={<Blogs />} />
      <Route path="/fa/blogs" element={<Blogs />} />
      <Route path="/fa/blogs/page/:pageNumber" element={<Blogs />} />
      <Route path="/fa/blogs/:slug" element={<BlogDetail />} />
      <Route path="/en/blogs" element={<Blogs />} />
      <Route path="/en/blogs/page/:pageNumber" element={<Blogs />} />
      <Route path="/en/blogs/:slug" element={<BlogDetail />} />
      <Route path="/ar/blogs" element={<Blogs />} />
      <Route path="/ar/blogs/page/:pageNumber" element={<Blogs />} />
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
    </Suspense>
  );
}
