import { Score, Text, TypingLanguage, User } from "@/types";
import { fetchApi } from "./fetch";

export const getAuthenticatedUser = async () => fetchApi<User>("/user");

export const logout = async () =>
  fetchApi<string>("/user/logout", { method: "POST" });

type TextQueryParams = {
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
    query: TextQueryParams;
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

  return fetchApi<Text | null>(`/users/${userId}/text${queryString}`, {
    headers,
  });
};

type TextParams = {
  specialCharacters: number;
  numbers: number;
  punctuation: boolean;
  language: TypingLanguage;
};

export const createNewText = async ({
  cookie,
  query,
}: {
  cookie?: string;
  query: TextParams;
}) => {
  // const body = Object.entries(query).reduce((acc, [key, value]) => {
  //   if (value === undefined) {
  //     return acc;
  //   }

  //   const queryParamKey = key.replace("Gte", "[gte]").replace("Lte", "[lte]");
  //   acc[queryParamKey] = value;

  //   return acc;
  // }, {} as { [key: string]: number | boolean | TypingLanguage });

  const headers = cookie ? { cookie } : undefined;

  return fetchApi<Text>(`/texts`, {
    headers,
    method: "POST",
    body: JSON.stringify(query),
  });
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
