import { api } from "./client";
import type { Category, Inquiry, Page, Product } from "../types";
export const catalogApi = {
  products: (q = "") => api<Page<Product>>(`/products${q ? `?${q}` : ""}`),
  product: (slug: string) => api<Product>(`/products/${slug}`),
  categories: () => api<Category[]>("/categories"),
  inquiry: (v: Partial<Inquiry>) =>
    api<Inquiry>("/inquiries", { method: "POST", body: JSON.stringify(v) }),
};
export const adminApi = {
  products: (q = "") =>
    api<Page<Product>>(`/admin/products${q ? `?${q}` : ""}`, {}, true),
  product: (id: string) => api<Product>(`/admin/products/${id}`, {}, true),
  saveProduct: (v: Partial<Product>, id?: string) =>
    api<Product>(
      `/admin/products${id ? `/${id}` : ""}`,
      { method: id ? "PUT" : "POST", body: JSON.stringify(v) },
      true,
    ),
  deleteProduct: (id: string) =>
    api(`/admin/products/${id}`, { method: "DELETE" }, true),
  deleteImage: (id: string, publicId: string) =>
    api<Product>(
      `/admin/products/${id}/images`,
      { method: "DELETE", body: JSON.stringify({ publicId }) },
      true,
    ),
  setPrimaryImage: (id: string, publicId: string) =>
    api<Product>(
      `/admin/products/${id}/images/primary`,
      { method: "PATCH", body: JSON.stringify({ publicId }) },
      true,
    ),
  reorderImages: (id: string, publicIds: string[]) =>
    api<Product>(
      `/admin/products/${id}/images/order`,
      { method: "PATCH", body: JSON.stringify({ publicIds }) },
      true,
    ),
  categories: () => api<Category[]>("/admin/categories", {}, true),
  saveCategory: (v: Partial<Category>, id?: string) =>
    api<Category>(
      `/admin/categories${id ? `/${id}` : ""}`,
      { method: id ? "PUT" : "POST", body: JSON.stringify(v) },
      true,
    ),
  deleteCategory: (id: string) =>
    api(`/admin/categories/${id}`, { method: "DELETE" }, true),
  inquiries: (q = "") =>
    api<Page<Inquiry>>(`/admin/inquiries${q ? `?${q}` : ""}`, {}, true),
  inquiry: (id: string) => api<Inquiry>(`/admin/inquiries/${id}`, {}, true),
  status: (id: string, status: string) =>
    api<Inquiry>(
      `/admin/inquiries/${id}/status`,
      { method: "PATCH", body: JSON.stringify({ status }) },
      true,
    ),
  upload: async (
    id: string,
    files: File[],
    options: { alt: string; isPrimary: boolean },
  ) => {
    const b = new FormData();
    files.forEach((file) => b.append("images", file));
    b.append("alt", options.alt);
    b.append("isPrimary", String(options.isPrimary));
    return api<Product>(
      `/admin/products/${id}/images`,
      { method: "POST", body: b },
      true,
    );
  },
};
