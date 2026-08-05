import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { useSearchParams } from "react-router-dom";
import { catalogApi } from "../../api/catalog";
import { Button, Field } from "../../components/ui";
import { inquirySchema, type InquiryValues } from "../../utils/validation";
export function InquiryPage() {
  const [p] = useSearchParams();
  const form = useForm<InquiryValues>({
    resolver: zodResolver(inquirySchema),
    defaultValues: {
      fullName: "",
      email: "",
      phone: "",
      message: "",
      productId: p.get("productId") || undefined,
    },
  });
  const send = useMutation({
    mutationFn: catalogApi.inquiry,
    onSuccess: () => form.reset(),
  });
  return (
    <section className="section">
      <div className="container form-page">
        <div>
          <p className="eyebrow">Pošaljite upit</p>
          <h1>Kako možemo da pomognemo?</h1>
          <p>
            Popunite formular i naš tim će vam se javiti sa potrebnim
            informacijama.
          </p>
        </div>
        <form onSubmit={form.handleSubmit((v) => send.mutate(v))} noValidate>
          {send.isSuccess && (
            <div className="alert success" role="status">
              Hvala. Vaš upit je uspešno poslat.
            </div>
          )}
          {send.isError && (
            <div className="alert error" role="alert">
              Upit nije poslat. Pokušajte ponovo.
            </div>
          )}
          <Field
            label="Ime i prezime"
            autoComplete="name"
            {...form.register("fullName")}
            error={form.formState.errors.fullName?.message}
          />
          <Field
            label="Email"
            type="email"
            autoComplete="email"
            {...form.register("email")}
            error={form.formState.errors.email?.message}
          />
          <Field
            label="Telefon"
            type="tel"
            autoComplete="tel"
            {...form.register("phone")}
            error={form.formState.errors.phone?.message}
          />
          <label className="field">
            <span>Poruka</span>
            <textarea rows={7} {...form.register("message")} />
            {form.formState.errors.message && (
              <small role="alert">
                {form.formState.errors.message.message}
              </small>
            )}
          </label>
          <Button type="submit" disabled={send.isPending}>
            {send.isPending ? "Slanje…" : "Pošaljite upit"}
          </Button>
          <p className="form-note">
            Vaše podatke koristimo isključivo radi odgovora na upit.
          </p>
        </form>
      </div>
    </section>
  );
}
