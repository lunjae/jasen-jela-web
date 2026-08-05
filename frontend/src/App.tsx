import { Route, Routes } from "react-router-dom";
import { AdminLayout } from "./layouts/AdminLayout";
import { PublicLayout } from "./layouts/PublicLayout";
import { CategoriesPage } from "./pages/admin/CategoriesPage";
import { DashboardPage } from "./pages/admin/DashboardPage";
import {
  InquiryDetailPage,
  InquiryListPage,
} from "./pages/admin/InquiriesAdmin";
import { LoginPage } from "./pages/admin/LoginPage";
import { ProductFormPage } from "./pages/admin/ProductFormPage";
import { ProductListPage } from "./pages/admin/ProductsAdmin";
import { CatalogPage } from "./pages/public/CatalogPage";
import { HomePage } from "./pages/public/HomePage";
import { InquiryPage } from "./pages/public/InquiryPage";
import { ProductPage } from "./pages/public/ProductPage";
import {
  AboutPage,
  ContactPage,
  NotFoundPage,
} from "./pages/public/StaticPages";
import { ProtectedRoute } from "./routes/ProtectedRoute";
export default function App() {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route index element={<HomePage />} />
        <Route path="proizvodi" element={<CatalogPage />} />
        <Route path="proizvodi/:slug" element={<ProductPage />} />
        <Route path="o-nama" element={<AboutPage />} />
        <Route path="kontakt" element={<ContactPage />} />
        <Route path="upit" element={<InquiryPage />} />
      </Route>
      <Route path="admin/prijava" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route path="admin" element={<AdminLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="proizvodi" element={<ProductListPage />} />
          <Route path="proizvodi/novi" element={<ProductFormPage />} />
          <Route path="proizvodi/:id" element={<ProductFormPage />} />
          <Route path="kategorije" element={<CategoriesPage />} />
          <Route path="upiti" element={<InquiryListPage />} />
          <Route path="upiti/:id" element={<InquiryDetailPage />} />
        </Route>
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
