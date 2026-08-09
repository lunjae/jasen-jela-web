import { renderToStaticMarkup } from "react-dom/server";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ProductCard } from "./ProductCard";
import type { Product } from "../types";
const product: Product = {
  id: "p1",
  name: "Mirni Jasen",
  slug: "mirni-jasen",
  shortDescription: "Klasičan model od jasena.",
  description: "Opis",
  categoryId: "c1",
  material: "Jasen",
  color: "Orah",
  currency: "RSD",
  images: [],
  featured: false,
  published: true,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};
describe("ProductCard", () => {
  it("shows the no-price state and accessible product link", () => {
    const html = renderToStaticMarkup(
      <MemoryRouter>
        <ProductCard product={product} />
      </MemoryRouter>,
    );
    expect(html).toContain("Cena na upit");
    expect(html).toContain("Detalji za Mirni Jasen");
  });
});
