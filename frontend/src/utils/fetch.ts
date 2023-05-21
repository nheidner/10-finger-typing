import { getApiUrl } from "./get_api_url";

export class FetchError extends Error {
  constructor(message: string, public status: number) {
    super(message);
  }
}

export const fetchApi = async <T>(url: string, options?: RequestInit) => {
  const baseUrl = getApiUrl();

  const res = await fetch(baseUrl + url, options);

  if (!res.ok) {
    throw new FetchError(res.statusText, res.status);
  }

  const data = (await res.json()) as { data: T };

  return data.data;
};
