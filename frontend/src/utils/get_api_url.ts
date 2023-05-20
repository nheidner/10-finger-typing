export const getApiUrl = () => {
  return typeof window === "undefined" ? "http://server_dev:8080/api" : "/api";
};
