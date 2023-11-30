import { useMutation } from "@tanstack/react-query";
import { FC, useEffect, useState } from "react";
import { startGame as startGameQuery } from "@/utils/queries";
import { GameStatus } from "@/types";

export const StartGameButton: FC<{
  gameStatus: GameStatus;
  roomId: string;
  setUserStartedGame: (userStartedGame: boolean) => void;
  userStartedGame: boolean;
}> = ({ gameStatus, roomId, setUserStartedGame, userStartedGame }) => {
  const { mutate: startGame } = useMutation({
    mutationFn: () => startGameQuery(roomId),
    mutationKey: ["start game"],
    onSuccess: () => {
      setUserStartedGame(true);
    },
  });

  const handleStartGame = () => {
    startGame();
  };

  let buttonText = "Start Game";
  if (gameStatus === "countdown") {
    buttonText = "Participate";
  }
  if (userStartedGame) {
    buttonText = "waiting for other users";
  }

  if (gameStatus !== "unstarted" && gameStatus !== "countdown") {
    return null;
  }

  return (
    <button
      disabled={userStartedGame}
      type="button"
      className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:bg-slate-500 disabled:hover:bg-slate-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
      onClick={handleStartGame}
    >
      {buttonText}
    </button>
  );
};
