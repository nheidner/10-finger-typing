/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  rewrites: async () => [
    {
      source: "/api/:path*",
      destination: `http://${process.env.BACKEND_HOST}:8080/api/:path*`,
    },
  ],
};

module.exports = nextConfig;
