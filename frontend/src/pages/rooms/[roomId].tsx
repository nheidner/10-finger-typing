import { CountDown } from "@/modules/room/components/Countdown";
import { GameDurationCounter } from "@/modules/room/components/GameDurationCounter";
import { RoomSubscriberList } from "@/modules/room/components/RoomSubscriberList";
import { Scores } from "@/modules/room/components/Scores";
import { StartGameButton } from "@/modules/room/components/StartGameButton";
import { useConnectToRoom } from "@/modules/room/hooks/use_connect_to_room";
import { Content } from "@/modules/train/components/Content";
import { GameStatus, Score } from "@/types";
import {
  createScore,
  getAuthenticatedUser,
  getTextById,
} from "@/utils/queries";
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
  const [userStartedGame, setUserStartedGame] = useState(false);
  const [scores, setScores] = useState<Score[]>([]);
  const { roomSubscribers, game, countDownDuration, roomSettings } =
    useConnectToRoom(roomId, setGameStatus, userStartedGame, setScores);

  const { data: textData, isLoading: textIsLoading } = useQuery(
    ["texts", game?.textId],
    () => getTextById(game?.textId!),
    {
      enabled: !!game?.textId,
    }
  );

  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: getAuthenticatedUser,
    retry: false,
  });

  const { mutate: createGameScore } = useMutation({
    mutationKey: ["create game score", roomId, game?.id],
    mutationFn: createScore,
  });

  const userStartedGameServerState = !!(
    authenticatedUserData?.id &&
    game?.gameSubscribers?.includes(authenticatedUserData.id)
  );

  useEffect(() => {
    if (gameStatus === "finished") {
      setUserStartedGame(false);
      return;
    }

    if (userStartedGameServerState && !userStartedGame) {
      setUserStartedGame(true);
    }
  }, [gameStatus, userStartedGameServerState, userStartedGame]);

  useEffect(() => {
    if (
      gameStatus === "finished" &&
      game?.textId &&
      roomSettings?.gameDurationSec &&
      game.id
    ) {
      createGameScore({
        roomId: roomId,
        body: {
          textId: game.textId,
          gameId: game.id,
          timeElapsed: roomSettings.gameDurationSec,
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
    roomSettings?.gameDurationSec,
    game?.id,
  ]);

  const isAdmin = roomSettings?.adminId === authenticatedUserData?.id;

  return (
    <>
      <Scores scores={scores} isAdmin={isAdmin} roomId={roomId} />
      <CountDown
        countDownDuration={countDownDuration}
        gameStatus={gameStatus}
        setGameStatus={setGameStatus}
      />
      <section className="flex justify-between items-center">
        <div className="flex gap-2 items-center">
          <RoomSubscriberList roomSubscribers={roomSubscribers} />
          <StartGameButton
            gameStatus={gameStatus}
            roomId={roomId}
            setUserStartedGame={setUserStartedGame}
            userStartedGame={userStartedGame}
          />
          <GameDurationCounter
            gameStatus={gameStatus}
            setGameStatus={setGameStatus}
            gameDuration={roomSettings?.gameDurationSec}
          />
        </div>
      </section>
      <Content
        isActive={gameStatus === "started" && userStartedGame}
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
