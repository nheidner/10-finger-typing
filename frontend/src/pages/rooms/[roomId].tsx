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
import { getTextById, startGame as startGameQuery } from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useMutation,
  useQuery,
} from "@tanstack/react-query";
import classNames from "classnames";
import { NextPage, NextPageContext } from "next";
import { useEffect, useRef, useState } from "react";

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
  type: "user_joined" | "user_left" | "initial_state" | "countdown_start";
  payload:
    | UserJoinedPayload
    | InitialStatePayload
    | UserLeftPayload
    | CountdownStartPayload;
};

const RoomPage: NextPage<{
  dehydratedState: DehydratedState;
  roomId: string;
}> = ({ roomId }) => {
  const [roomSubscribers, setRoomSubscribers] = useState<RoomSubscriber[]>([]);
  const [game, setGame] = useState<Game | null>(null);
  const [gameStatus, setGameStatus] = useState<GameStatus>("unstarted");

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

  const webSocketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!roomId) {
      return;
    }

    const apiUrl = getWsUrl();

    const websocketUrl = `${apiUrl}/rooms/${roomId}/ws`;
    webSocketRef.current = new WebSocket(websocketUrl);

    webSocketRef.current.onopen = () => {};

    webSocketRef.current.onmessage = (e) => {
      const message = JSON.parse(e.data) as Message;

      switch (message.type) {
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

          setTimeout(() => {
            setGameStatus("started");
          }, payload * 1000);
          break;
        }
        default:
          break;
      }

      console.log("message received: >>", JSON.parse(e.data));
    };

    webSocketRef.current.onclose = (e) => {
      console.log("connection closed: >>", e);
    };

    return () => {
      console.log("cloooosed");
      if (webSocketRef.current) {
        webSocketRef.current?.close(1000, "user left the room");
      }
    };
  }, [roomId]);

  return (
    <>
      <div>{roomId}</div>
      <section className="flex gap-2 items-center">
        {roomSubscribers
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
          })
          .map((roomSubscriber) => {
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
