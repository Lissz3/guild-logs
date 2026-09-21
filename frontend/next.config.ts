import type { NextConfig } from "next";

// The Go API. The browser only talks to Next.js; /api/* is proxied to Go.
const api = process.env.GUILDLOGS_API_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${api}/api/:path*` }];
  },
  experimental: {
    // The first analysis of a report downloads every event and can take minutes.
    proxyTimeout: 5 * 60 * 1000,
  },
};

export default nextConfig;
