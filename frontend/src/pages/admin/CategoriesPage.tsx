import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { adminApi } from "../../api/catalog";
import { Button, Modal, Spinner, State } from "../../components/ui";
import type { Category } from "../../types";
export function CategoriesPage() {
  const qc = useQueryClient();
  const [edit, setEdit] = useState<Partial<Category> | null>(null);
  const cats = useQuery({
    queryKey: ["admin-categories"],
    queryFn: adminApi.categories,
  });
  const save = useMutation({
    mutationFn: (c: Partial<Category>) => adminApi.saveCategory(c, c.id),
    onSuccess: () => {
      setEdit(null);
      qc.invalidateQueries({ queryKey: ["admin-categories"] });
    },
  });
  const del = useMutation({
    mutationFn: adminApi.deleteCategory,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin-categories"] }),
  });
  return (
    <>
      <div className="admin-heading">
        <div>
          <p className="eyebrow">Organizacija</p>
          <h1>Kategorije</h1>
        </div>
        <Button
          onClick={() =>
            setEdit({ name: "", description: "", published: true })
          }
        >
          Nova kategorija
        </Button>
      </div>
      {cats.isLoading ? (
        <Spinner />
      ) : (
        <div className="admin-panel table-wrap">
          <table>
            <thead>
              <tr>
                <th>Naziv</th>
                <th>Opis</th>
                <th>Status</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {cats.data?.map((c) => (
                <tr key={c.id}>
                  <td>
                    <strong>{c.name}</strong>
                    <small>{c.slug}</small>
                  </td>
                  <td>{c.description || "—"}</td>
                  <td>{c.published ? "Objavljena" : "Sakrivena"}</td>
                  <td className="row-actions">
                    <button onClick={() => setEdit(c)}>Uredi</button>
                    <button
                      onClick={() =>
                        confirm(`Obrisati kategoriju ${c.name}?`) &&
                        del.mutate(c.id)
                      }
                    >
                      Obriši
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!cats.data?.length && <State title="Još nema kategorija" />}
        </div>
      )}
      {edit && (
        <Modal
          title={edit.id ? "Uredite kategoriju" : "Nova kategorija"}
          onClose={() => setEdit(null)}
        >
          <form
            onSubmit={(e) => {
              e.preventDefault();
              save.mutate(edit);
            }}
          >
            <label className="field">
              <span>Naziv</span>
              <input
                value={edit.name || ""}
                onChange={(e) => setEdit({ ...edit, name: e.target.value })}
                required
              />
            </label>
            <label className="field">
              <span>Opis</span>
              <textarea
                rows={4}
                value={edit.description || ""}
                onChange={(e) =>
                  setEdit({ ...edit, description: e.target.value })
                }
              />
            </label>
            <label className="check">
              <input
                type="checkbox"
                checked={edit.published ?? true}
                onChange={(e) =>
                  setEdit({ ...edit, published: e.target.checked })
                }
              />{" "}
              Objavljena
            </label>
            <div className="actions">
              <button type="button" onClick={() => setEdit(null)}>
                Odustani
              </button>
              <Button disabled={save.isPending}>Sačuvajte</Button>
            </div>
          </form>
        </Modal>
      )}
    </>
  );
}
