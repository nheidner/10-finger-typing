import { GameStatus } from "@/types";
import { FC, useEffect, useState } from "react";

export const GameDurationCounter: FC<{
  gameStatus: GameStatus;
  setGameStatus: (newGameStatus: GameStatus) => void;
  gameDuration?: number;
}> = ({ gameStatus, setGameStatus, gameDuration }) => {
  const [remainingTime, setRemainingTime] = useState(gameDuration);

  useEffect(() => {
    if (gameStatus !== "started" || !remainingTime) {
      return;
    }

    const timeout = setTimeout(() => {
      setRemainingTime(
        (oldRemainingTime) => oldRemainingTime && oldRemainingTime - 1
      );
    }, 1000);

    return () => clearTimeout(timeout);
  }, [gameStatus, remainingTime, setGameStatus]);

  useEffect(() => {
    if (remainingTime === 0) {
      setGameStatus("finished");
      setRemainingTime(gameDuration);
      return;
    }
  }, [remainingTime, gameDuration, setGameStatus]);

  useEffect(() => {
    if (gameDuration) {
      setRemainingTime(gameDuration);
    }
  }, [gameDuration]);

  if (gameStatus !== "started") {
    return null;
  }

  return <div className="text-xl">{remainingTime}</div>;
};
