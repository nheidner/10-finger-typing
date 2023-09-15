export const getApiUrl = () => {
  return typeof window === "undefined" ? "http://server_dev:8080/api" : "/api";
};

export const getWsUrl = () => {
  return typeof window === "undefined"
    ? "ws://server_dev:8080/api"
    : "ws://localhost:8080/api";
};
