import { FC, Fragment, useState } from "react";
import { Dialog, Transition } from "@headlessui/react";
import { InvitePanel } from "./InvitePanel";
import { useRouter } from "next/router";
import { User } from "@/types";
import { Modal } from "@/components/Modal";

export const InviteModal: FC<{
  isOpen: boolean;
  setOpen: (open: boolean) => void;
}> = ({ isOpen, setOpen }) => {
  const [newRoomUsers, setNewRoomUsers] = useState<Partial<User>[]>([]);

  const router = useRouter();
  const { textId } = router.query as {
    textId?: string;
  };

  if (!textId) {
    return null;
  }

  const addNewRoomUser = (user: Partial<User>) => {
    setNewRoomUsers((users) => users.concat(user));
  };

  const removeNewRoomUser = (idx: number) => {
    setNewRoomUsers((prevUsers) => prevUsers.filter((_, i) => i !== idx));
  };

  const removeNewRoomUsers = () => {
    setNewRoomUsers([]);
  };

  const closeModal = () => setOpen(false);

  const panelHeight = `${
    (152 + Math.ceil(newRoomUsers.length / 3) * 148) / 16
  }rem`;

  return (
    <Modal isOpen={isOpen} setOpen={setOpen} panelHeight={panelHeight}>
      <InvitePanel
        textId={textId}
        closeModal={closeModal}
        addNewRoomUser={addNewRoomUser}
        removeNewRoomUser={removeNewRoomUser}
        removeNewRoomUsers={removeNewRoomUsers}
        newRoomUsers={newRoomUsers}
      />
    </Modal>
  );
};
