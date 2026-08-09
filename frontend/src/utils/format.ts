export const money = (value: number, currency: string) =>
  new Intl.NumberFormat("sr-Latn-RS", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(value);
export const date = (value: string) =>
  new Intl.DateTimeFormat("sr-Latn-RS", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
