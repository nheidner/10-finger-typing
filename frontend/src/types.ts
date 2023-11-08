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

export type GameStatus = "unstarted" | "started" | "finished";
export type SubscriberGameStatus = "unstarted" | "started" | "finished";
export type SubscriberStatus = "inactive" | "active";
