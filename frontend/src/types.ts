export type Uuid = string;

export interface User {
  id: Uuid;
  username: string;
  firstName: string;
  email: string;
  lastName: string;
  isVerified: boolean;
}

export interface Score {
  id: Uuid;
  wordsPerMinute: number;
  wordsTyped: number;
  timeElapsed: number;
  accuracy: number;
  numberErrors: number;
  errors: { [error: string]: number };
  userId: Uuid;
}

export type LanguageCode = (typeof languageCodes)[number];
export type LanguageName = (typeof languageNames)[number];

export interface Text {
  id: Uuid;
  language: LanguageCode;
  text: string;
  punctuation: boolean;
  specialCharacters: number;
  numbers: number;
}

export interface Game {
  id: Uuid;
  roomId: Uuid;
  status: GameStatus;
  textId: Uuid;
  gameSubscribers: string[] | null;
}

export type GameStatus = "unstarted" | "started" | "finished" | "countdown";
export type SubscriberGameStatus = "unstarted" | "started" | "finished";
export type SubscriberStatus = "inactive" | "active";

export interface RoomSubscriber {
  userId: Uuid;
  gameStatus: SubscriberGameStatus;
  status: SubscriberStatus;
  username: string;
}

export interface Room {
  adminId: Uuid;
  currentGame: Game;
  roomSubscribers: RoomSubscriber[];
  gameDurationSec: number;
  currentGameScores: Score[];
}

type RoomInvitationPayload = {
  by: string;
  roomId: Uuid;
};

export interface UserNotification {
  id: Uuid;
  type: "room_invitation";
  payload: RoomInvitationPayload;
}
