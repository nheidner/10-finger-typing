import { Score, TypingLanguage, User } from "@/types";
import { fetchApi } from "./fetch";

export const getAuthenticatedUser = async () => fetchApi<User>("/user");

export const logout = async () =>
  fetchApi<string>("/user/logout", { method: "POST" });

type TextQuery = {
  specialCharactersGte?: number;
  specialCharactersLte?: number;
  numbersGte?: number;
  numbersLte?: number;
  punctuation?: boolean;
  language: TypingLanguage;
};

export const getNewTextByUserid = async (
  userId: number,
  {
    cookie,
    query,
  }: {
    cookie?: string;
    query: TextQuery;
  }
) => {
  const queryString = Object.entries(query).reduce(
    (acc, [key, value], index) => {
      if (value === undefined) {
        return acc;
      }

      const queryParamDelimiter = index === 0 ? "?" : "&";
      const queryParamKey = key.replace("Gte", "[gte]").replace("Lte", "[lte]");

      return `${acc}${queryParamDelimiter}${queryParamKey}=${encodeURIComponent(
        value
      )}`;
    },
    ""
  );

  const headers = cookie ? { cookie } : undefined;

  return fetchApi<Text>(`/users/${userId}/text${queryString}`, { headers });
};

type UserCredentials = {
  email: string;
  password: string;
};

export const login = async (userCredentials: UserCredentials) =>
  fetchApi<User>("/user/login", {
    method: "POST",
    body: JSON.stringify(userCredentials),
  });

export const getUserByUsername = async (username: string, cookie?: string) => {
  const queryParams = `username=${encodeURIComponent(username)}`;
  const queryString = `?${queryParams}`;

  const headers = cookie ? { cookie } : undefined;

  const users = await fetchApi<User[]>(`/users${queryString}`, { headers });
  return users[0];
};

export const getScoresByUsername = async (
  username: string,
  { cookie, sortBy }: { cookie?: string; sortBy?: string[] }
) => {
  const queryParams = sortBy
    ?.map((sortByValue) => `sort_by=${encodeURIComponent(sortByValue)}`)
    .concat(`username=${encodeURIComponent(username)}`)
    .join("&");
  const queryString = queryParams ? `?${queryParams}` : "";

  const headers = cookie ? { cookie } : undefined;

  return fetchApi<Score[]>(`/scores${queryString}`, {
    headers,
  });
};
