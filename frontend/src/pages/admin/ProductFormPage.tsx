import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import { adminApi } from "../../api/catalog";
import { Button, Spinner } from "../../components/ui";
import type { Product } from "../../types";
type Form = Omit<
  Product,
  "id" | "slug" | "images" | "createdAt" | "updatedAt"
> & { price?: number };
const defaults: Form = {
  name: "",
  shortDescription: "",
  description: "",
  categoryId: "",
  material: "",
  color: "",
  currency: "RSD",
  featured: false,
  published: false,
};
export function ProductFormPage() {
  const { id } = useParams();
  const nav = useNavigate();
  const qc = useQueryClient();
  const [files, setFiles] = useState<File[]>([]);
  const existing = useQuery({
    queryKey: ["admin-product", id],
    queryFn: () => adminApi.product(id!),
    enabled: !!id,
  });
  const cats = useQuery({
    queryKey: ["admin-categories"],
    queryFn: adminApi.categories,
  });
  const f = useForm<Form>({ defaultValues: defaults });
  useEffect(() => {
    if (existing.data) {
      const { ...x } = existing.data;
      f.reset(x);
    }
  }, [existing.data, f]);
  const save = useMutation({
    mutationFn: async (v: Form) => {
      const p = await adminApi.saveProduct(
        { ...v, images: existing.data?.images || [] },
        id,
      );
      if (files.length) await adminApi.upload(p.id, files);
      return p;
    },
    onSuccess: (p) => {
      qc.invalidateQueries({ queryKey: ["admin-products"] });
      nav(`/admin/proizvodi/${p.id}`, { replace: true });
    },
  });
  const imageAction = useMutation({
    mutationFn: async (action: {
      type: "primary" | "delete";
      path: string;
    }) => {
      if (!id || !existing.data) throw new Error("Product is not saved");
      if (action.type === "delete")
        return adminApi.deleteImage(id, action.path);
      const images = existing.data.images.map((image) => ({
        ...image,
        isPrimary: image.storagePath === action.path,
      }));
      return adminApi.saveProduct({ ...existing.data, images }, id);
    },
    onSuccess: (product) => {
      qc.setQueryData(["admin-product", id], product);
      qc.invalidateQueries({ queryKey: ["admin-products"] });
    },
  });
  if (id && existing.isLoading) return <Spinner />;
  return (
    <>
      <div className="admin-heading">
        <div>
          <p className="eyebrow">Proizvodi</p>
          <h1>{id ? "Uredite proizvod" : "Novi proizvod"}</h1>
        </div>
      </div>
      <form
        className="admin-form"
        onSubmit={f.handleSubmit((v) => save.mutate(v))}
      >
        <section className="admin-panel">
          <h2>Osnovni podaci</h2>
          <div className="form-grid">
            <label className="field wide">
              <span>Naziv</span>
              <input {...f.register("name", { required: true })} />
            </label>
            <label className="field wide">
              <span>Kratak opis</span>
              <input
                {...f.register("shortDescription", {
                  required: true,
                  minLength: 10,
                })}
              />
            </label>
            <label className="field wide">
              <span>Detaljan opis</span>
              <textarea
                rows={7}
                {...f.register("description", {
                  required: true,
                  minLength: 20,
                })}
              />
            </label>
            <label className="field">
              <span>Kategorija</span>
              <select {...f.register("categoryId", { required: true })}>
                <option value="">Izaberite</option>
                {cats.data?.map((c) => (
                  <option value={c.id} key={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Materijal</span>
              <input {...f.register("material", { required: true })} />
            </label>
            <label className="field">
              <span>Boja</span>
              <input {...f.register("color", { required: true })} />
            </label>
            <label className="field">
              <span>Cena</span>
              <input
                type="number"
                min="0"
                step="1"
                {...f.register("price", {
                  valueAsNumber: true,
                  setValueAs: (v) => (Number.isNaN(v) ? undefined : v),
                })}
              />
            </label>
            <label className="field">
              <span>Valuta</span>
              <select {...f.register("currency")}>
                <option>RSD</option>
                <option>EUR</option>
              </select>
            </label>
          </div>
        </section>
        <section className="admin-panel">
          <h2>Dimenzije (cm)</h2>
          <div className="form-grid thirds">
            <label className="field">
              <span>Dužina</span>
              <input
                type="number"
                {...f.register("dimensions.length", { valueAsNumber: true })}
              />
            </label>
            <label className="field">
              <span>Širina</span>
              <input
                type="number"
                {...f.register("dimensions.width", { valueAsNumber: true })}
              />
            </label>
            <label className="field">
              <span>Visina</span>
              <input
                type="number"
                {...f.register("dimensions.height", { valueAsNumber: true })}
              />
            </label>
          </div>
        </section>
        <section className="admin-panel">
          <h2>Slike</h2>
          {existing.data?.images.length ? (
            <div className="admin-images">
              {existing.data.images.map((x) => (
                <div key={x.storagePath}>
                  <img src={x.url} alt={x.alt} />
                  <button
                    type="button"
                    disabled={x.isPrimary || imageAction.isPending}
                    onClick={() =>
                      imageAction.mutate({
                        type: "primary",
                        path: x.storagePath,
                      })
                    }
                  >
                    {x.isPrimary ? "Glavna" : "Postavi kao glavnu"}
                  </button>
                  <button
                    type="button"
                    disabled={imageAction.isPending}
                    onClick={() =>
                      confirm("Ukloniti ovu sliku?") &&
                      imageAction.mutate({
                        type: "delete",
                        path: x.storagePath,
                      })
                    }
                  >
                    Ukloni
                  </button>
                </div>
              ))}
            </div>
          ) : null}
          <label className="upload">
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp"
              multiple
              onChange={(e) => setFiles(Array.from(e.target.files || []))}
            />
            <span>Izaberite do 8 slika (najviše 5 MB po slici)</span>
          </label>
        </section>
        <section className="admin-panel publish">
          <label>
            <input type="checkbox" {...f.register("featured")} /> Istaknuti
            proizvod
          </label>
          <label>
            <input type="checkbox" {...f.register("published")} /> Objavljen u
            katalogu
          </label>
        </section>
        {save.isError && (
          <div className="alert error">
            Čuvanje nije uspelo. Proverite unete podatke.
          </div>
        )}
        <div className="form-actions">
          <button type="button" onClick={() => nav("/admin/proizvodi")}>
            Odustani
          </button>
          <Button disabled={save.isPending}>
            {save.isPending ? "Čuvanje…" : "Sačuvajte proizvod"}
          </Button>
        </div>
      </form>
    </>
  );
}
