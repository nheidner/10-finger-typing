import { Modal } from "@/components/Modal";
import { ScoresTable } from "@/components/ScoresTable";
import { SelectNewText } from "@/modules/room/components/SelectNewText";
import { Score } from "@/types";
import { FC } from "react";

export const Scores: FC<{
  scores: Score[];
  isAdmin: boolean;
  roomId: string;
}> = ({ scores, isAdmin, roomId }) => {
  const selectNewText = isAdmin ? <SelectNewText roomId={roomId} /> : null;

  return (
    <Modal isOpen={scores.length > 0} setIsOpen={() => {}}>
      {selectNewText}
      <ScoresTable scores={scores} />
    </Modal>
  );
};
