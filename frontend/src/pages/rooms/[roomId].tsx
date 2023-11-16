import { Avatar } from "@/components/Avatar";
import { Content } from "@/modules/train/components/Content";
import {
  Game,
  GameStatus,
  SubscriberGameStatus,
  SubscriberStatus,
  User,
} from "@/types";
import { getWsUrl } from "@/utils/get_api_url";
import {
  getAuthenticatedUser,
  getTextById,
  startGame as startGameQuery,
} from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useMutation,
  useQuery,
} from "@tanstack/react-query";
import classNames from "classnames";
import { NextPage, NextPageContext } from "next";
import { useEffect, useMemo, useRef, useState } from "react";

interface RoomSubscriber {
  userId: string;
  gameStatus: SubscriberGameStatus;
  status: SubscriberStatus;
  username: string;
}

type InitialStatePayload = {
  adminId: string;
  currentGame: Game;
  roomSubscribers: RoomSubscriber[];
};

type UserJoinedPayload = string;

type UserLeftPayload = string;

type CountdownStartPayload = number;

type Message = {
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

const notReceivePongReason = "did not receive pong";
const apiUrl = getWsUrl();

const RoomPage: NextPage<{
  dehydratedState: DehydratedState;
  roomId: string;
}> = ({ roomId }) => {
  const [roomSubscribers, setRoomSubscribers] = useState<RoomSubscriber[]>([]);
  const [game, setGame] = useState<Game | null>(null);
  const [gameStatus, setGameStatus] = useState<GameStatus>("unstarted");
  const [count, setCount] = useState<number | null>(null);

  const { data: textData, isLoading: textIsLoading } = useQuery(
    ["texts", game?.textId],
    () => getTextById(game?.textId!),
    {
      enabled: !!game?.textId,
    }
  );

  const { mutate: startGame } = useMutation({
    mutationFn: () => startGameQuery(roomId),
    mutationKey: ["start game"],
  });

  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: getAuthenticatedUser,
    retry: false,
  });

  const websocketRef = useRef<WebSocket | null>(null);

  const sortedRoomSubscribers = useMemo(() => {
    return roomSubscribers
      .filter((roomSubscriber) => {
        return roomSubscriber.userId != authenticatedUserData?.id;
      })
      ?.sort((roomSubscriberA, roomSubscriberB) => {
        const roomSubscriberAStatusA =
          roomSubscriberA.status === "inactive"
            ? "inactive"
            : roomSubscriberA.gameStatus;
        const roomSubscriberAStatusB =
          roomSubscriberB.status === "inactive"
            ? "inactive"
            : roomSubscriberB.gameStatus;

        return (
          statusToSortingWeightMapping[roomSubscriberAStatusB] -
          statusToSortingWeightMapping[roomSubscriberAStatusA]
        );
      });
  }, [roomSubscribers, authenticatedUserData?.id]);

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

            setCount(payload);
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
  }, [roomId]);

  useEffect(() => {
    if (count) {
      const timer = setTimeout(() => {
        setCount((oldCount) => {
          if (oldCount) {
            return oldCount - 1;
          }

          return oldCount;
        });
      }, 1000);

      return () => clearTimeout(timer);
    }

    if (count === 0) {
      setTimeout(() => {
        setGameStatus("started");
        setCount(null);
      }, 200);
    }
  }, [count]);

  const countDown =
    count !== null ? (
      <div className="fixed text-9xl left-1/2 -translate-x-1/2 top-1/2 -translate-y-1/2">
        {count > 0 ? count : "start"}
      </div>
    ) : null;

  return (
    <>
      {countDown}
      <div>{roomId}</div>
      <section className="flex gap-2 items-center">
        {sortedRoomSubscribers.map((roomSubscriber) => {
          const user: Partial<User> = {
            id: roomSubscriber.userId,
            username: roomSubscriber.username,
          };

          const status =
            roomSubscriber.status === "inactive"
              ? "inactive"
              : roomSubscriber.gameStatus;

          const color = statusToColorMapping[status];

          return (
            <div
              key={roomSubscriber.userId}
              style={{ borderColor: color }}
              className={classNames(
                "rounded-full",
                color ? "border-2" : "p-[2px]"
              )}
            >
              <Avatar
                user={user}
                textClassName="text-1xl"
                containerClassName="w-10 h-10"
              />
            </div>
          );
        })}
        <button
          type="button"
          className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
          onClick={() => startGame()}
        >
          Start Game
        </button>
      </section>
      <Content
        isActive={gameStatus == "started"}
        isLoading={textIsLoading}
        text={textData || null}
        userData={{}}
        onType={() => {}}
      />
    </>
  );
};

const statusToSortingWeightMapping: {
  [key in SubscriberGameStatus | "inactive"]: number;
} = {
  inactive: 0,
  unstarted: 1,
  started: 2,
  finished: 3,
};
const statusToColorMapping: {
  [key in SubscriberGameStatus | "inactive"]: string | undefined;
} = {
  inactive: undefined,
  unstarted: "#aaa",
  started: "#77f",
  finished: "#7f7",
};

RoomPage.getInitialProps = async (ctx: NextPageContext) => {
  const roomId = ctx.query.roomId as string;

  const queryClient = new QueryClient();

  return {
    dehydratedState: dehydrate(queryClient),
    roomId,
  };
};

export default RoomPage;
