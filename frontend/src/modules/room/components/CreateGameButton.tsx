import { GameStatus } from "@/types";
import { createGame } from "@/utils/queries";
import { useMutation } from "@tanstack/react-query";
import { FC } from "react";

export const CreateNewGameButton: FC<{
  gameStatus: GameStatus;
  roomId: string;
  textId?: string;
}> = ({ gameStatus, roomId, textId }) => {
  const { mutate: createNewGame } = useMutation({
    mutationKey: ["create new game"],
    mutationFn: createGame,
  });

  if (!textId) {
    return null;
  }

  const handleStartGame = () => {
    createNewGame({
      roomId,
      body: {
        textId,
      },
    });
  };

  if (gameStatus !== "finished") {
    return null;
  }

  return (
    <button
      type="button"
      onClick={handleStartGame}
      className="inline-flex items-center  rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:bg-slate-500 disabled:hover:bg-slate-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
    >
      create new game
    </button>
  );
};
