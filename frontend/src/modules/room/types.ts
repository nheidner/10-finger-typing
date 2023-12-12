import { Game, Room, Score, User, Uuid } from "@/types";

export type InitialStatePayload = Room;
export type UserJoinedPayload = Uuid;
export type UserLeftPayload = Uuid;
export type CountdownStartPayload = number;
export type NewGamePayload = Game;
export type GameResultPayload = Score[];
export type GameStartedPayload = null;
export type CursorPayload = {
  position: number;
  userId: Uuid;
};
export type UserStartedGamePayload = Uuid;
export type UserFinishedGamePayload = Uuid;

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
    | "game_started"
    | "cursor"
    | "user_started_game"
    | "user_finished_game";
  payload:
    | UserJoinedPayload
    | InitialStatePayload
    | UserLeftPayload
    | CountdownStartPayload
    | NewGamePayload
    | GameResultPayload
    | GameStartedPayload
    | CursorPayload
    | UserStartedGamePayload
    | UserFinishedGamePayload;
};
