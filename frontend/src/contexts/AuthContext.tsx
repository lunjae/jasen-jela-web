/* oxlint-disable react/only-export-components -- provider and its matching hook intentionally share one private context */
import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import {
  onAuthStateChanged,
  signInWithEmailAndPassword,
  signOut,
  type User,
} from "firebase/auth";
import { auth } from "../config/firebase";
type Value = {
  user: User | null;
  loading: boolean;
  login: (e: string, p: string) => Promise<void>;
  logout: () => Promise<void>;
};
const C = createContext<Value | null>(null);
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(
    () =>
      onAuthStateChanged(auth, (u) => {
        setUser(u);
        setLoading(false);
      }),
    [],
  );
  return (
    <C.Provider
      value={{
        user,
        loading,
        login: async (e, p) => {
          await signInWithEmailAndPassword(auth, e, p);
        },
        logout: () => signOut(auth),
      }}
    >
      {children}
    </C.Provider>
  );
}
export function useAuth() {
  const v = useContext(C);
  if (!v) throw new Error("AuthProvider is missing");
  return v;
}
