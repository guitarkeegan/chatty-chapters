/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/v1/:path*", // Proxy to API
      },
    ];
  },
};

export default nextConfig;
