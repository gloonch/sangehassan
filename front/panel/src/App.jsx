import AppRoutes from "./router";
import { AuthProvider } from "./lib/auth";

export default function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}
