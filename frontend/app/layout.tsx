import type { Metadata } from "next";
import Script from "next/script";
import "./globals.css";

export const metadata: Metadata = {
  title: "Guild Logs · Raid analysis",
  description: "Warcraft Logs report analysis: avoidable deaths, avoidable damage, activity and defensives.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        {/* Wowhead's tooltip widget: turns any wowhead.com/spell=<id> link into a
            real ability tooltip (description, cooldown, etc.) on hover. WCL's API
            only gives us name/icon, not tooltip text, so this fills that gap. */}
        <Script id="wowhead-tooltips-config" strategy="beforeInteractive">
          {"window.whTooltips = { colorLinks: false, iconizeLinks: false, renameLinks: false };"}
        </Script>
        <Script src="https://wow.zamimg.com/js/tooltips.js" strategy="afterInteractive" />
      </body>
    </html>
  );
}
