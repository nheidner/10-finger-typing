import { GameStatus } from "../types";
import { FC, useEffect, useState } from "react";

const defaultGameDuration = 30;

export const GameDurationCounter: FC<{
  gameStatus: GameStatus;
  setGameStatus: (newGameStatus: GameStatus) => void;
  gameDuration: null | number;
}> = ({ gameStatus, setGameStatus, gameDuration }) => {
  const [remainingTime, setRemainingTime] = useState(
    gameDuration || defaultGameDuration
  );

  useEffect(() => {
    if (gameStatus !== "started") {
      return;
    }
    if (remainingTime === 0) {
      setGameStatus("finished");
      return;
    }

    const timeout = setTimeout(() => {
      setRemainingTime((oldRemainingTime) => oldRemainingTime - 1);
    }, 1000);

    return () => clearTimeout(timeout);
  }, [gameStatus, remainingTime, setGameStatus]);

  if (gameStatus !== "started") {
    return null;
  }

  return <div className="text-xl">{remainingTime}</div>;
};
