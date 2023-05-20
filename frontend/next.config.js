/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  rewrites: async () => [
    {
      source: "/api/:path*",
      destination: "http://server_dev:8080/api/:path*",
    },
  ],
};

module.exports = nextConfig;
