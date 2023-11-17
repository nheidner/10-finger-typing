import { CountDown } from "@/modules/room/components/Countdown";
import { GameDurationCounter } from "@/modules/room/components/GameDurationCounter";
import { RoomSubscriberList } from "@/modules/room/components/RoomSubscriberList";
import { StartGameButton } from "@/modules/room/components/StartGameButton";
import { useConnectToRoom } from "@/modules/room/hooks/use_connect_to_room";
import { GameStatus } from "@/modules/room/types";
import { Content } from "@/modules/train/components/Content";
import { getTextById } from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage, NextPageContext } from "next";
import { useState } from "react";

const RoomPage: NextPage<{
  dehydratedState: DehydratedState;
  roomId: string;
}> = ({ roomId }) => {
  const [gameStatus, setGameStatus] = useState<GameStatus>("unstarted");

  const { roomSubscribers, game, countDownDuration, gameDuration } =
    useConnectToRoom(roomId, setGameStatus);

  const { data: textData, isLoading: textIsLoading } = useQuery(
    ["texts", game?.textId],
    () => getTextById(game?.textId!),
    {
      enabled: !!game?.textId,
    }
  );

  return (
    <>
      <CountDown
        countDownDuration={countDownDuration}
        gameStatus={gameStatus}
        setGameStatus={setGameStatus}
      />
      <section className="flex gap-2 items-center">
        <RoomSubscriberList roomSubscribers={roomSubscribers} />
        <StartGameButton gameStatus={gameStatus} roomId={roomId} />
        <GameDurationCounter
          gameStatus={gameStatus}
          setGameStatus={setGameStatus}
          gameDuration={gameDuration}
        />
      </section>
      <Content
        isActive={gameStatus === "started"}
        isLoading={textIsLoading}
        text={textData || null}
        userData={{}}
        onType={() => {}}
      />
    </>
  );
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
