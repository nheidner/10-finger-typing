import { Game, Room, Score, User } from "@/types";

export type InitialStatePayload = Room;
export type UserJoinedPayload = string;
export type UserLeftPayload = string;
export type CountdownStartPayload = number;
export type NewGamePayload = Game;
export type GameResultPayload = Score[];
export type GameStartedPayload = null;

export type Message = {
  user: User;
  type:
    | "user_joined"
    | "user_left"
    | "initial_state"
    | "countdown"
    | "pong"
    | "new_game"
    | "game_result"
    | "game_started";
  payload:
    | UserJoinedPayload
    | InitialStatePayload
    | UserLeftPayload
    | CountdownStartPayload
    | NewGamePayload
    | GameResultPayload
    | GameStartedPayload;
};
