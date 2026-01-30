import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import ProductsLanding from "./pages/ProductsLanding";
import Products from "./pages/Products";
import ProductDetail from "./pages/ProductDetail";
import Blogs from "./pages/Blogs";
import Login from "./pages/Login";
import Signup from "./pages/Signup";
import Profile from "./pages/Profile";
import StoneFinishes from "./pages/StoneFinishes";
import Gallery from "./pages/Gallery";
import About from "./pages/About";
import Contact from "./pages/Contact";
import BlocksLanding from "./pages/BlocksLanding";
import BlocksCatalog from "./pages/BlocksCatalog";
import BlockDetail from "./pages/BlockDetail";

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/products/overview" element={<ProductsLanding />} />
      <Route path="/products" element={<Products />} />
      <Route path="/products/:slug" element={<ProductDetail />} />
      <Route path="/blocks" element={<BlocksLanding />} />
      <Route path="/blocks/catalog" element={<BlocksCatalog />} />
      <Route path="/blocks/:slug" element={<BlockDetail />} />
      <Route path="/blogs" element={<Blogs />} />
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Signup />} />
      <Route path="/profile" element={<Profile />} />
      <Route path="/stone-finishes" element={<StoneFinishes />} />
      <Route path="/gallery" element={<Gallery />} />
      <Route path="/about" element={<About />} />
      <Route path="/contact" element={<Contact />} />
    </Routes>
  );
}
