export interface User {
  id: string;
  username: string;
  firstName: string;
  email: string;
  lastName: string;
  isVerified: boolean;
}

export interface Score {
  id: string;
  wordsPerMinute: number;
  wordsTyped: number;
  timeElapsed: number;
  accuracy: number;
  numberErrors: number;
  errors: { [error: string]: number };
  userId: string;
}

export type LanguageCode = (typeof languageCodes)[number];
export type LanguageName = (typeof languageNames)[number];

export interface Text {
  id: string;
  language: LanguageCode;
  text: string;
  punctuation: boolean;
  specialCharacters: number;
  numbers: number;
}

export interface Game {
  id: string;
  roomId: string;
  status: GameStatus;
  textId: string;
}

export interface Room {
  id: string;
}

type RoomInvitationPayload = {
  by: string;
  roomId: string;
};

export interface UserNotification {
  id: string;
  type: "room_invitation";
  payload: RoomInvitationPayload;
}

export type GameStatus = "unstarted" | "started" | "finished";
export type SubscriberGameStatus = "unstarted" | "started" | "finished";
export type SubscriberStatus = "inactive" | "active";
