export interface ProductImage {
  url: string;
  storagePath: string;
  alt: string;
  order: number;
  isPrimary: boolean;
}
export interface Product {
  id: string;
  name: string;
  slug: string;
  shortDescription: string;
  description: string;
  categoryId: string;
  material: string;
  color: string;
  dimensions?: { length?: number; width?: number; height?: number };
  price?: number;
  currency: "RSD" | "EUR";
  images: ProductImage[];
  featured: boolean;
  published: boolean;
  createdAt: string;
  updatedAt: string;
}
export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  published: boolean;
  createdAt: string;
  updatedAt: string;
}
export type InquiryStatus = "new" | "read" | "contacted" | "closed";
export interface Inquiry {
  id: string;
  productId?: string;
  fullName: string;
  email: string;
  phone: string;
  message: string;
  status: InquiryStatus;
  createdAt: string;
  updatedAt: string;
}
export interface Page<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}
export interface ApiError {
  code: string;
  message: string;
  fields?: Record<string, string>;
}
