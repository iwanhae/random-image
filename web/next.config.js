/** @type {import('next').NextConfig} */
module.exports = {
  reactStrictMode: true,
  rewrites: async () => {
    return [
      {
        source: "/api/sample",
        destination: "http://127.0.0.1:8080/api/sample",
      },
      {
        source: "/api/group/:id*",
        destination: "http://127.0.0.1:8080/api/group/:id*",
      },
      {
        source: "/data/:id*",
        destination: "http://127.0.0.1:8080/data/:id*",
      },
    ]
  }
}
