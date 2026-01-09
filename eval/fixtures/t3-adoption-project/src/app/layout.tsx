import "@/styles/globals.css";
import { Inter } from "next/font/google";
import { TRPCProvider } from "@/components/providers/trpc-provider";
import { AuthProvider } from "@/components/providers/auth-provider";

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });

export const metadata = {
  title: "T3 Blog App",
  description: "A blog application built with the T3 Stack",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`font-sans ${inter.variable}`}>
        <AuthProvider>
          <TRPCProvider>
            <main className="min-h-screen bg-gradient-to-b from-slate-900 to-slate-800">
              {children}
            </main>
          </TRPCProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
