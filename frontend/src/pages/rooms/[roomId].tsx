import { Avatar } from "@/components/Avatar";
import { Content } from "@/modules/train/components/Content";
import { Game, SubscriberGameStatus, SubscriberStatus, User } from "@/types";
import { getWsUrl } from "@/utils/get_api_url";
import { getTextById } from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
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

type Message = {
  user: User;
  type: "user_joined" | "user_left" | "initial_state";
  payload: UserJoinedPayload | InitialStatePayload | UserLeftPayload;
};

const RoomPage: NextPage<{
  dehydratedState: DehydratedState;
  roomId: string;
}> = ({ roomId }) => {
  const [roomSubscribers, setRoomSubscribers] = useState<RoomSubscriber[]>([]);
  const [game, setGame] = useState<Game | null>(null);

  const { data: textData, isLoading: textIsLoading } = useQuery(
    ["texts", game?.textId],
    () => getTextById(game?.textId!),
    {
      enabled: !!game?.textId,
    }
  );

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
        default:
          break;
      }

      console.log("message received: >>", JSON.parse(e.data));
    };

    webSocketRef.current.onclose = (e) => {
      console.log("connection closed: >>", e);
    };

    return () => {
      if (webSocketRef.current) {
        webSocketRef.current?.close(1000, "user left the room");
      }
    };
  }, [roomId]);

  return (
    <>
      <div>{roomId}</div>
      <section className="flex gap-2 items-center">
        {roomSubscribers?.map((roomSubscriber) => {
          const user: Partial<User> = {
            id: roomSubscriber.userId,
            username: roomSubscriber.username,
          };

          const color = getColor(
            roomSubscriber.gameStatus,
            roomSubscriber.status
          );

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
      </section>
      <Content
        isLoading={textIsLoading}
        text={textData || null}
        userData={{}}
        onType={() => {}}
      />
    </>
  );
};

const getColor = (
  subscriberGameStatus: SubscriberGameStatus,
  status: SubscriberStatus
): string | undefined => {
  switch (subscriberGameStatus) {
    case "unstarted":
      if (status == "active") {
        return "#aaa";
      }

      return undefined;
    case "started":
      return "#77f";
    case "finished":
      return "#7f7";
  }
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
