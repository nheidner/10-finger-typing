import { User } from "@/types";

export interface Game {
  id: string;
  roomId: string;
  status: GameStatus;
  textId: string;
}

export type GameStatus = "unstarted" | "started" | "finished" | "countdown";
export type SubscriberGameStatus = "unstarted" | "started" | "finished";
export type SubscriberStatus = "inactive" | "active";

export interface RoomSubscriber {
  userId: string;
  gameStatus: SubscriberGameStatus;
  status: SubscriberStatus;
  username: string;
}

export type InitialStatePayload = {
  adminId: string;
  currentGame: Game;
  roomSubscribers: RoomSubscriber[];
};

export type UserJoinedPayload = string;

export type UserLeftPayload = string;

export type CountdownStartPayload = number;

export type Message = {
  user: User;
  type:
    | "user_joined"
    | "user_left"
    | "initial_state"
    | "countdown_start"
    | "pong";
  payload:
    | UserJoinedPayload
    | InitialStatePayload
    | UserLeftPayload
    | CountdownStartPayload;
};
