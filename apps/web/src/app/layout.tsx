import type { Metadata } from "next";
import { Geist, Press_Start_2P } from "next/font/google";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";
import { cn } from "@/lib/utils";

const geistSans = Geist({
  subsets: ["latin"],
  variable: "--font-sans",
});

const pressStart2P = Press_Start_2P({
  subsets: ["latin"],
  weight: "400",
  variable: "--font-display",
});

export const metadata: Metadata = {
  title: "Polina",
  description: "Editor colaborativo de missões",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="pt-BR"
      className={cn(
        "h-full antialiased font-sans",
        geistSans.variable,
        pressStart2P.variable
      )}
    >
      <body className="flex min-h-full flex-col">
        {children}
        <Toaster richColors />
      </body>
    </html>
  );
}
