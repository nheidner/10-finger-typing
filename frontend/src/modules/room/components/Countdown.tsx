import { FC, useEffect, useState } from "react";
import { GameStatus } from "@/types";

export const CountDown: FC<{
  gameStatus: GameStatus;
  setGameStatus: (newGameStatus: GameStatus) => void;
  countDownDuration: null | number;
}> = ({ setGameStatus, gameStatus, countDownDuration }) => {
  const [count, setCount] = useState<number | null>(null);

  useEffect(() => {
    if (count) {
      const timer = setTimeout(() => {
        setCount((oldCount) => {
          if (oldCount) {
            return oldCount - 1;
          }

          return oldCount;
        });
      }, 1000);

      return () => clearTimeout(timer);
    }

    if (count === 0) {
      setTimeout(() => {
        setGameStatus("started");
        setCount(null);
      }, 200);
    }
  }, [count, setGameStatus]);

  useEffect(() => {
    if (gameStatus === "countdown" && countDownDuration) {
      setCount(countDownDuration);
    }
  }, [gameStatus, countDownDuration]);

  if (count === null) {
    return null;
  }

  return (
    <div className="fixed text-9xl left-1/2 -translate-x-1/2 top-1/2 -translate-y-1/2">
      {count > 0 ? count : "start"}
    </div>
  );
};
