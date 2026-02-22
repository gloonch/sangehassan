import { Navigate, Route, Routes } from "react-router-dom";
import Home from "./pages/Home";
import ProductsLanding from "./pages/ProductsLanding";
import ProductDetail from "./pages/ProductDetail";
import Blogs from "./pages/Blogs";
import Login from "./pages/Login";
import Signup from "./pages/Signup";
import Profile from "./pages/Profile";
import Gallery from "./pages/Gallery";
import About from "./pages/About";
import BlocksLanding from "./pages/BlocksLanding";
import BlocksCatalog from "./pages/BlocksCatalog";
import BlockDetail from "./pages/BlockDetail";
import Ads from "./pages/Ads";
import AdDetail from "./pages/AdDetail";
import NewAd from "./pages/NewAd";
import RequireUserAuth from "./components/RequireUserAuth";

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/products" element={<ProductsLanding />} />
      <Route path="/products/overview" element={<Navigate to="/products" replace />} />
      <Route path="/products/:slug" element={<ProductDetail />} />
      <Route path="/blocks" element={<BlocksLanding />} />
      <Route path="/blocks/catalog" element={<BlocksCatalog />} />
      <Route path="/blocks/:slug" element={<BlockDetail />} />
      <Route element={<RequireUserAuth />}>
        <Route path="/ads" element={<Ads />} />
        <Route path="/ads/new" element={<NewAd />} />
        <Route path="/ads/:id" element={<AdDetail />} />
      </Route>
      <Route path="/blogs" element={<Blogs />} />
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Signup />} />
      <Route path="/profile" element={<Profile />} />
      <Route path="/gallery" element={<Gallery />} />
      <Route path="/about" element={<About />} />
    </Routes>
  );
}
