/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://server:8080/api/:path*'
      }
    ]
  },
  reactStrictMode: true,
}

module.exports = nextConfig
