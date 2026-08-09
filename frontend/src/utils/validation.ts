import { z } from "zod";
export const inquirySchema = z.object({
  fullName: z.string().trim().min(2, "Unesite ime i prezime.").max(120),
  email: z.email("Unesite ispravnu email adresu."),
  phone: z
    .string()
    .trim()
    .regex(/^[+0-9][0-9 ()/-]{5,24}$/, "Unesite ispravan broj telefona."),
  message: z
    .string()
    .trim()
    .min(10, "Poruka mora imati najmanje 10 znakova.")
    .max(3000),
  productId: z.string().optional(),
});
export type InquiryValues = z.infer<typeof inquirySchema>;
export const loginSchema = z.object({
  email: z.email("Unesite ispravnu email adresu."),
  password: z.string().min(6, "Lozinka mora imati najmanje 6 znakova."),
});
