import { getWsUrl } from "@/utils/get_api_url";
import { useEffect, useRef, useState } from "react";
import {
  CountdownStartPayload,
  Game,
  GameStatus,
  InitialStatePayload,
  Message,
  RoomSubscriber,
  UserJoinedPayload,
  UserLeftPayload,
} from "../types";

const notReceivePongReason = "did not receive pong";
const apiUrl = getWsUrl();

export const useConnectToRoom = (
  roomId: string,
  setGameStatus: (newGameStatus: GameStatus) => void
) => {
  const [roomSubscribers, setRoomSubscribers] = useState<RoomSubscriber[]>([]);
  const [game, setGame] = useState<Game | null>(null);
  const [countDownDuration, setCountDownDuration] = useState<null | number>(
    null
  );
  const [gameDuration, setGameDuration] = useState<null | number>(null);

  const websocketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let retryTimes = 0;
    const baseDelay = 1000;
    const maxDelay = 8000;
    let pingInterval: ReturnType<typeof setInterval> | null = null;
    let waitForPongTimeout: ReturnType<typeof setTimeout> | null = null;
    let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
    let isComponentMounted = true;

    const connectSocket = () => {
      retryTimes = 0;
      const websocketConnectionIsAlreadyInUse =
        websocketRef.current?.readyState === WebSocket.OPEN ||
        websocketRef.current?.readyState === WebSocket.CONNECTING;

      if (!roomId || websocketConnectionIsAlreadyInUse) {
        return;
      }

      const websocketUrl = `${apiUrl}/rooms/${roomId}/ws`;

      websocketRef.current = new WebSocket(websocketUrl);

      websocketRef.current.onopen = () => {
        pingInterval = setInterval(() => {
          if (websocketRef.current?.readyState === WebSocket.OPEN) {
            const ping = {
              type: "ping",
            };

            websocketRef.current.send(JSON.stringify(ping));

            if (!waitForPongTimeout) {
              waitForPongTimeout = setTimeout(() => {
                websocketRef.current?.close(1000, notReceivePongReason);
              }, 1000);
            }
          }
        }, 2000);
      };

      websocketRef.current.onmessage = (e) => {
        if (!isComponentMounted) {
          return;
        }

        const message = JSON.parse(e.data) as Message;

        switch (message.type) {
          case "pong": {
            if (waitForPongTimeout) {
              clearTimeout(waitForPongTimeout);
              waitForPongTimeout = null;
            }

            break;
          }
          case "initial_state": {
            const payload = message.payload as InitialStatePayload;

            setGame(payload.currentGame);
            setRoomSubscribers(payload.roomSubscribers);
            break;
          }
          case "user_joined": {
            const payload = message.payload as UserJoinedPayload;

            setRoomSubscribers((oldRoomSubscribers) =>
              oldRoomSubscribers.map((roomSubscriber) => {
                if (roomSubscriber.userId !== payload) {
                  return roomSubscriber;
                }

                return {
                  ...roomSubscriber,
                  status: "active",
                };
              })
            );
            break;
          }
          case "user_left": {
            const payload = message.payload as UserLeftPayload;

            setRoomSubscribers((oldRoomSubscribers) =>
              oldRoomSubscribers.map((roomSubscriber) => {
                if (roomSubscriber.userId !== payload) {
                  return roomSubscriber;
                }

                return {
                  ...roomSubscriber,
                  status: "inactive",
                };
              })
            );
            break;
          }
          case "countdown_start": {
            const payload = message.payload as CountdownStartPayload;

            setGameStatus("countdown");
            setCountDownDuration(payload);
            setGameDuration(10);
            break;
          }
          default:
            break;
        }
      };

      websocketRef.current.onclose = (e) => {
        if (waitForPongTimeout) {
          clearTimeout(waitForPongTimeout);
          waitForPongTimeout = null;
        }
        if (pingInterval) {
          clearInterval(pingInterval);
          pingInterval = null;
        }
        if (e.wasClean) {
          console.log(
            `Connection closed cleanly, code=${e.code}, reason=${e.reason}`
          );
          if (e.reason == notReceivePongReason) {
            const backoffDelay = Math.min(
              baseDelay * 2 ** retryTimes + Math.random() * 1000,
              maxDelay
            );
            retryTimes++;

            if (!reconnectTimeout) {
              reconnectTimeout = setTimeout(() => {
                connectSocket();
                reconnectTimeout = null;
              }, backoffDelay);
            }
          }

          return;
        }

        console.error("Connection died", e);
        // Here you can implement reconnection logic if needed.
      };
    };

    connectSocket();

    return () => {
      isComponentMounted = false;

      if (websocketRef.current) {
        websocketRef.current.close(1000, "user left the room");
      }
      if (pingInterval) {
        clearInterval(pingInterval);
      }
      if (waitForPongTimeout) {
        clearTimeout(waitForPongTimeout);
      }
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
    };
  }, [roomId, setGameStatus]);

  return { roomSubscribers, game, countDownDuration, gameDuration };
};
