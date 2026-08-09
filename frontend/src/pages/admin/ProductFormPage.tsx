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
  const [previews, setPreviews] = useState<string[]>([]);
  const [imageAlt, setImageAlt] = useState("");
  const [uploadAsPrimary, setUploadAsPrimary] = useState(false);
  const [imageError, setImageError] = useState("");
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
  useEffect(() => {
    if (!files.length) {
      setPreviews([]);
      return;
    }
    const urls = files.map((file) => URL.createObjectURL(file));
    setPreviews(urls);
    return () => urls.forEach((url) => URL.revokeObjectURL(url));
  }, [files]);
  function selectImages(next: File[]) {
    upload.reset();
    setImageError("");
    if (!next.length) {
      setFiles([]);
      return;
    }
    if (next.length > 8) {
      setImageError("Možete izabrati najviše 8 slika odjednom.");
      setFiles([]);
      return;
    }
    if (next.some((file) => !["image/jpeg", "image/png", "image/webp"].includes(file.type))) {
      setImageError("Dozvoljeni formati su JPG, PNG i WebP.");
      setFiles([]);
      return;
    }
    if (next.some((file) => file.size > 10 * 1024 * 1024)) {
      setImageError("Svaka slika može imati najviše 10 MB.");
      setFiles([]);
      return;
    }
    setFiles(next);
    if (!imageAlt) setImageAlt(f.getValues("name"));
  }
  const save = useMutation({
    mutationFn: async (v: Form) => {
      const p = await adminApi.saveProduct(v, id);
      return files.length
        ? adminApi.upload(p.id, files, {
            alt: imageAlt,
            isPrimary: uploadAsPrimary,
          })
        : p;
    },
    onSuccess: (p) => {
      setFiles([]);
      setImageError("");
      qc.invalidateQueries({ queryKey: ["admin-products"] });
      nav(`/admin/proizvodi/${p.id}`, { replace: true });
    },
  });
  const upload = useMutation({
    mutationFn: () => {
      if (!id || !files.length) throw new Error("Izaberite slike.");
      return adminApi.upload(id, files, {
        alt: imageAlt,
        isPrimary: uploadAsPrimary,
      });
    },
    onSuccess: (product) => {
      qc.setQueryData(["admin-product", id], product);
      qc.invalidateQueries({ queryKey: ["admin-products"] });
      setFiles([]);
      setImageAlt("");
      setUploadAsPrimary(false);
      setImageError("");
    },
    onError: () => setImageError("Upload slike nije uspeo."),
  });
  const imageAction = useMutation({
    mutationFn: async (action: {
      type: "primary" | "delete";
      path: string;
    }) => {
      if (!id || !existing.data) throw new Error("Product is not saved");
      if (action.type === "delete")
        return adminApi.deleteImage(id, action.path);
      return adminApi.setPrimaryImage(id, action.path);
    },
    onSuccess: (product) => {
      qc.setQueryData(["admin-product", id], product);
      qc.invalidateQueries({ queryKey: ["admin-products"] });
    },
  });
  const reorder = useMutation({
    mutationFn: (publicIds: string[]) => adminApi.reorderImages(id!, publicIds),
    onSuccess: (product) => {
      qc.setQueryData(["admin-product", id], product);
      qc.invalidateQueries({ queryKey: ["admin-products"] });
    },
  });
  const images = [...(existing.data?.images || [])].sort(
    (a, b) => a.order - b.order,
  );
  function moveImage(index: number, offset: number) {
    const target = index + offset;
    if (!id || target < 0 || target >= images.length) return;
    const next = [...images];
    [next[index], next[target]] = [next[target], next[index]];
    reorder.mutate(next.map((image) => image.publicId));
  }
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
          {images.length ? (
            <div className="admin-images">
              {images.map((x, index) => (
                <div key={x.publicId}>
                  <img src={x.url} alt={x.alt} />
                  <p>{x.alt || "Bez alternativnog teksta"}</p>
                  {x.isPrimary && <strong>Glavna slika</strong>}
                  <button
                    type="button"
                    disabled={x.isPrimary || imageAction.isPending}
                    onClick={() =>
                      imageAction.mutate({
                        type: "primary",
                        path: x.publicId,
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
                        path: x.publicId,
                      })
                    }
                  >
                    Ukloni
                  </button>
                  <button
                    type="button"
                    disabled={index === 0 || reorder.isPending}
                    onClick={() => moveImage(index, -1)}
                    aria-label="Pomeri sliku ulevo"
                  >
                    ←
                  </button>
                  <button
                    type="button"
                    disabled={index === images.length - 1 || reorder.isPending}
                    onClick={() => moveImage(index, 1)}
                    aria-label="Pomeri sliku udesno"
                  >
                    →
                  </button>
                </div>
              ))}
            </div>
          ) : null}
          <label
            className="upload"
            onDragOver={(event) => event.preventDefault()}
            onDrop={(event) => {
              event.preventDefault();
              selectImages(Array.from(event.dataTransfer.files));
            }}
          >
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp"
              multiple
              onChange={(event) => selectImages(Array.from(event.target.files || []))}
            />
            <span>Izaberite ili prevucite do 8 slika (najviše 10 MB po slici)</span>
          </label>
          {previews.length > 0 && (
            <div className="upload-preview">
              <div className="upload-preview-images">
                {previews.map((preview, index) => (
                  <img key={preview} src={preview} alt={`Pregled slike ${index + 1}`} />
                ))}
              </div>
              <label className="field">
                <span>Alternativni tekst</span>
                <input
                  value={imageAlt}
                  maxLength={200}
                  onChange={(event) => setImageAlt(event.target.value)}
                />
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={uploadAsPrimary}
                  onChange={(event) => setUploadAsPrimary(event.target.checked)}
                />{" "}
                Postavi kao glavnu sliku
              </label>
              {id && (
                <Button
                  type="button"
                  disabled={upload.isPending}
                  onClick={() => upload.mutate()}
                >
                  {upload.isPending ? "Otpremanje…" : "Otpremi sliku"}
                </Button>
              )}
            </div>
          )}
          {!id && files.length > 0 && (
            <p className="form-note">Slike će biti otpremljene nakon čuvanja proizvoda.</p>
          )}
          {imageError && <div className="alert error">{imageError}</div>}
          {upload.isSuccess && <div className="alert success">Slika je uspešno otpremljena.</div>}
          {(imageAction.isError || reorder.isError) && (
            <div className="alert error">Izmena slika nije uspela.</div>
          )}
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
