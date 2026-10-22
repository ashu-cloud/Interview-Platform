import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  cacheComponents: true,
  partialPrefetching: true,
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
  async redirects() {
    return [
      {
        source: "/dashboard",
        destination: "/console",
        permanent: true,
      },
      {
        source: "/login",
        destination: "/auth",
        permanent: true,
      },
      {
        source: "/reports/:id",
        destination: "/report/:id",
        permanent: true,
      },
      {
        source: "/roles/new",
        destination: "/blueprints/new",
        permanent: true,
      },
    ];
  },
};

export default nextConfig;
