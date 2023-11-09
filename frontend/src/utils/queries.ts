import { Room, Score, Text, LanguageCode, User } from "@/types";
import { fetchApi } from "./fetch";

export type NewRoomBodyParams = {
  userIds: string[];
  emails: string[];
};

export const createRoom = async ({ body }: { body: NewRoomBodyParams }) => {
  return fetchApi<Room>("/rooms", {
    method: "POST",
    body: JSON.stringify(body),
  });
};

export type NewGameBodyParams = { textId: string };

export const createGame = async ({
  roomId,
  body,
}: {
  roomId: string;
  body: NewGameBodyParams;
}) => {
  return fetchApi<{ id: string }>(`/rooms/${roomId}/games`, {
    method: "POST",
    body: JSON.stringify(body),
  });
};

export const createRoomAndText = async ({
  newGameBody,
  newRoomBody,
}: {
  newRoomBody: NewRoomBodyParams;
  newGameBody: NewGameBodyParams;
}) => {
  const room = await createRoom({ body: newRoomBody });
  const { id: gameId } = await createGame({
    roomId: room.id,
    body: newGameBody,
  });

  return {
    room,
    gameId,
  };
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
  language: LanguageCode;
};

export const getTextById = async (textId: string) => {
  return fetchApi<Text>(`/texts/${textId}`);
};

export const startGame = async (roomId: string) => {
  return fetchApi<string>(`/rooms/${roomId}/start_game`, { method: "POST" });
};

export const getNewTextByUserid = async (
  userId: string,
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

  return fetchApi<Text>(`/users/${userId}/text${queryString}`, {
    headers,
  });
};

type TextParams = {
  specialCharacters: number;
  numbers: number;
  punctuation: boolean;
  language: LanguageCode;
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
