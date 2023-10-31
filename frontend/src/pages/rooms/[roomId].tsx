import { User } from "@/types";
import { getWsUrl } from "@/utils/get_api_url";
import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage, NextPageContext } from "next";
import { useEffect, useRef } from "react";

type Message = {
  user: User;
  type: "user_added" | "cursor";
  payload: { cursor: number } | null;
};

const RoomPage: NextPage<{
  dehydratedState: DehydratedState;
  roomId: string;
}> = ({ roomId }) => {
  const webSocketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!roomId) {
      return;
    }

    const apiUrl = getWsUrl();

    const websocketUrl = `${apiUrl}/rooms/${roomId}/ws`;
    webSocketRef.current = new WebSocket(websocketUrl);

    webSocketRef.current.onopen = () => {
      const message = {
        type: "user_added",
      };
      webSocketRef.current?.send(JSON.stringify(message));
    };

    webSocketRef.current.onmessage = (e) => {
      const message = JSON.parse(e.data) as Message;
      console.log("message :>> ", message);

      switch (message.type) {
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

  return <div>{roomId}</div>;
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
