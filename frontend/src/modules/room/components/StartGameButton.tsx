import { useMutation } from "@tanstack/react-query";
import { FC, useEffect, useState } from "react";
import { startGame as startGameQuery } from "@/utils/queries";
import { GameStatus } from "@/types";

export const StartGameButton: FC<{
  gameStatus: GameStatus;
  roomId: string;
}> = ({ gameStatus, roomId }) => {
  const [userMarkedGameStart, setUserMarkedGameStart] = useState(false);

  const { mutate: startGame } = useMutation({
    mutationFn: () => startGameQuery(roomId),
    mutationKey: ["start game"],
  });

  useEffect(() => {
    if (gameStatus === "unstarted") {
      setUserMarkedGameStart(false);
    }
  }, [gameStatus]);

  const handleStartGame = () => {
    setUserMarkedGameStart(true);
    startGame();
  };

  let buttonText = "Start Game";
  if (gameStatus === "countdown") {
    buttonText = "Participate";
  }
  if (userMarkedGameStart) {
    buttonText = "waiting for other users";
  }

  if (gameStatus !== "unstarted" && gameStatus !== "countdown") {
    return null;
  }

  return (
    <button
      disabled={userMarkedGameStart}
      type="button"
      className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:bg-slate-500 disabled:hover:bg-slate-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
      onClick={handleStartGame}
    >
      {buttonText}
    </button>
  );
};
