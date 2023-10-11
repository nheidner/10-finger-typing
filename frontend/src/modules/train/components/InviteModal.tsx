import { FC, Fragment, useState } from "react";
import { Dialog, Transition } from "@headlessui/react";
import { InvitePanel } from "./InvitePanel";
import { useRouter } from "next/router";
import { User } from "@/types";

export const InviteModal: FC<{
  isOpen: boolean;
  setOpen: (open: boolean) => void;
}> = ({ isOpen, setOpen }) => {
  const [newRoomUsers, setNewRoomUsers] = useState<Partial<User>[]>([]);

  const router = useRouter();
  const { textId } = router.query as {
    textId?: string;
  };

  const addNewRoomUser = (user: Partial<User>) => {
    setNewRoomUsers((users) => users.concat(user));
  };

  const removeNewRoomUser = (idx: number) => {
    setNewRoomUsers((prevUsers) => prevUsers.filter((_, i) => i !== idx));
  };

  const removeNewRoomUsers = () => {
    setNewRoomUsers([]);
  };

  if (!textId) {
    return null;
  }

  const closeModal = () => setOpen(false);

  return (
    <Transition.Root show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-10" onClose={setOpen}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-50"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-out duration-20"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black bg-opacity-25 transition-opacity backdrop-blur-sm" />
        </Transition.Child>
        <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <InvitePanel
              textId={textId}
              closeModal={closeModal}
              addNewRoomUser={addNewRoomUser}
              removeNewRoomUser={removeNewRoomUser}
              removeNewRoomUsers={removeNewRoomUsers}
              newRoomUsers={newRoomUsers}
            />
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
};
