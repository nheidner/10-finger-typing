import { Room, Score, Text, TypingLanguage, User } from "@/types";
import { fetchApi } from "./fetch";

export type NewRoomParams = {
  userIds: number[];
  emails: string[];
  textIds: number[];
};

export const createRoom = async ({ query }: { query: NewRoomParams }) => {
  return fetchApi<Room>("/rooms", {
    method: "POST",
    body: JSON.stringify(query),
  });
};

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

export const getUsersByUsernamePartial = async (
  usernamePartial: string,
  cookie?: string
) => {
  if (usernamePartial === "") {
    return Promise.resolve([]);
  }

  const queryString = `?username_contains=${encodeURIComponent(
    usernamePartial
  )}`;

  const headers = cookie ? { cookie } : undefined;

  return fetchApi<User[]>(`/users${queryString}`, { headers });
};

export const getScoresByUsername = async (
  username: string,
  { cookie, sortBy }: { cookie?: string; sortBy?: string[] }
) => {
  const sortByQueryParams = sortBy?.map(
    (sortByValue) => `sort_by=${encodeURIComponent(sortByValue)}`
  );

  const queryParams = [
    `username=${encodeURIComponent(username)}`,
    ...(sortByQueryParams || []),
  ].join("&");

  const queryString = queryParams ? `?${queryParams}` : "";

  const headers = cookie ? { cookie } : undefined;

  return fetchApi<Score[]>(`/scores${queryString}`, {
    headers,
  });
};
