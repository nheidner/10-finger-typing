import { Game, Room, User } from "@/types";

export type InitialStatePayload = Room;
export type UserJoinedPayload = string;
export type UserLeftPayload = string;
export type CountdownStartPayload = number;
export type NewGamePayload = Game;

export type Message = {
  user: User;
  type:
    | "user_joined"
    | "user_left"
    | "initial_state"
    | "countdown_start"
    | "pong"
    | "new_game";
  payload:
    | UserJoinedPayload
    | InitialStatePayload
    | UserLeftPayload
    | CountdownStartPayload
    | NewGamePayload;
};
