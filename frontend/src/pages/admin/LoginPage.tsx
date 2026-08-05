import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { z } from "zod";
import { Button, Field } from "../../components/ui";
import { useAuth } from "../../contexts/AuthContext";
import { loginSchema } from "../../utils/validation";
type Values = z.infer<typeof loginSchema>;
export function LoginPage() {
  const { user, login } = useAuth();
  const nav = useNavigate();
  const loc = useLocation();
  const [error, setError] = useState("");
  const f = useForm<Values>({ resolver: zodResolver(loginSchema) });
  if (user) return <Navigate to="/admin" replace />;
  return (
    <main className="login-page">
      <form
        onSubmit={f.handleSubmit(async (v) => {
          try {
            setError("");
            await login(v.email, v.password);
            nav(
              (loc.state as { from?: { pathname: string } })?.from?.pathname ||
                "/admin",
            );
          } catch {
            setError("Prijava nije uspela. Proverite podatke.");
          }
        })}
      >
        <div className="login-brand">JJ</div>
        <h1>Administracija</h1>
        <p>Prijavite se administratorskim nalogom.</p>
        {error && <div className="alert error">{error}</div>}
        <Field
          label="Email"
          type="email"
          {...f.register("email")}
          error={f.formState.errors.email?.message}
        />
        <Field
          label="Lozinka"
          type="password"
          {...f.register("password")}
          error={f.formState.errors.password?.message}
        />
        <Button className="full" disabled={f.formState.isSubmitting}>
          {f.formState.isSubmitting ? "Prijava…" : "Prijavite se"}
        </Button>
      </form>
    </main>
  );
}
