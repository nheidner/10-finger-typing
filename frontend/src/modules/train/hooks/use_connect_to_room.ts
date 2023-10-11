import { Text, User } from "@/types";
import { getWsUrl } from "@/utils/get_api_url";
import { SetStateAction, useEffect, useRef } from "react";
import { UserData } from "../types";

type Message = {
  user: User;
  type: "user_added" | "cursor";
  payload: { cursor: number } | null;
};

export const useConnectToRoom = (
  setUserData: (
    value: SetStateAction<{
      [userId: number]: UserData;
    }>
  ) => void,
  roomId?: string,
  textData?: Text
) => {
  const webSocketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!textData?.id || !roomId) {
      return;
    }

    const apiUrl = getWsUrl();

    const websocketUrl = `${apiUrl}/rooms/${roomId}/ws?textId=${textData.id}`;
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
        case "user_added":
          setUserData((userData) => {
            return {
              ...userData,
              [message.user.id]: { user: message.user, cursor: 0 },
            };
          });
          break;
        case "cursor":
          setUserData((userData) => {
            return {
              ...userData,
              [message.user.id]: {
                ...userData[message.user.id],
                cursor: message.payload?.cursor || 0,
              },
            };
          });
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
  }, [textData?.id, roomId, setUserData]);

  return webSocketRef;
};
