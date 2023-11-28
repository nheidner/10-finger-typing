import { CountDown } from "@/modules/room/components/Countdown";
import { CreateNewGameButton } from "@/modules/room/components/CreateGameButton";
import { GameDurationCounter } from "@/modules/room/components/GameDurationCounter";
import { RoomSubscriberList } from "@/modules/room/components/RoomSubscriberList";
import { StartGameButton } from "@/modules/room/components/StartGameButton";
import { useConnectToRoom } from "@/modules/room/hooks/use_connect_to_room";
import { Content } from "@/modules/train/components/Content";
import { GameStatus } from "@/types";
import { createScore, getTextById } from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useMutation,
  useQuery,
} from "@tanstack/react-query";
import { NextPage, NextPageContext } from "next";
import { useEffect, useState } from "react";

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

  const { mutate: createGameScore } = useMutation({
    mutationKey: ["create game score", roomId, game?.id],
    mutationFn: createScore,
  });

  useEffect(() => {
    if (gameStatus === "finished" && game?.textId && gameDuration && game.id) {
      createGameScore({
        roomId: roomId,
        body: {
          textId: game.textId,
          gameId: game.id,
          timeElapsed: gameDuration,
          errors: { a: 5, b: 3, c: 1 },
          wordsTyped: 40,
        },
      });
    }
  }, [
    gameStatus,
    game?.textId,
    createGameScore,
    roomId,
    gameDuration,
    game?.id,
  ]);

  return (
    <>
      <CountDown
        countDownDuration={countDownDuration}
        gameStatus={gameStatus}
        setGameStatus={setGameStatus}
      />
      <section className="flex justify-between items-center">
        <div className="flex gap-2 items-center">
          <RoomSubscriberList roomSubscribers={roomSubscribers} />
          <StartGameButton gameStatus={gameStatus} roomId={roomId} />
          <GameDurationCounter
            gameStatus={gameStatus}
            setGameStatus={setGameStatus}
            gameDuration={gameDuration}
          />
        </div>
        <CreateNewGameButton
          gameStatus={gameStatus}
          roomId={roomId}
          textId={game?.textId}
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
