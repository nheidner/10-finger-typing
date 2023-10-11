import { FC, Fragment, useMemo, useState } from "react";
import { Dialog, Transition } from "@headlessui/react";
import { User } from "@/types";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  NewRoomParams,
  createRoom,
  getAuthenticatedUser,
} from "@/utils/queries";
import { useRouter } from "next/router";
import { UserAutocompleteBox } from "./UserAutocompleteBox";
import { UserList } from "./UserList";

const sanitizeNewRoomUsers = (
  users: Partial<User>[],
  textId: string
): NewRoomParams => {
  const emails: string[] = [];
  const userIds: number[] = [];

  for (const user of users) {
    if (user.id) {
      userIds.push(user.id);
      continue;
    }

    if (user.email) {
      emails.push(user.email);
    }
  }

  return {
    emails,
    userIds,
    textIds: [parseInt(textId)],
  };
};

export const InvitePanel: FC<{
  textId: string;
  closeModal: () => void;
  newRoomUsers: Partial<User>[];
  addNewRoomUser: (user: Partial<User>) => void;
  removeNewRoomUser: (idx: number) => void;
  removeNewRoomUsers: () => void;
}> = ({
  textId,
  closeModal,
  newRoomUsers,
  addNewRoomUser,
  removeNewRoomUser,
  removeNewRoomUsers,
}) => {
  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: getAuthenticatedUser,
    retry: false,
  });

  const router = useRouter();

  const { mutate: createNewRoomMutate, isLoading } = useMutation({
    mutationKey: ["create room"],
    mutationFn: createRoom,
    onSuccess: (data) => {
      router.push({
        pathname: router.pathname,
        query: { ...router.query, roomId: data.id },
      });

      removeNewRoomUsers();

      closeModal();
    },
  });

  const handleCreateNewRoom = () => {
    const params = sanitizeNewRoomUsers(newRoomUsers, textId);

    createNewRoomMutate({ query: params });
  };

  const newRoomUsersDisplaySet = useMemo(
    () =>
      new Set(
        newRoomUsers.map(
          (newRoomUser) => (newRoomUser.username || newRoomUser.email)!
        )
      )
        .add(authenticatedUserData!.email)
        .add(authenticatedUserData!.username),
    [newRoomUsers, authenticatedUserData]
  );

  return (
    <Transition.Child
      as={Fragment}
      enter="ease-out duration-50"
      enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
      enterTo="opacity-100 translate-y-0 sm:scale-100"
      leave="ease-out duration-20"
      leaveFrom="opacity-100 translate-y-0 sm:scale-100"
      leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
    >
      <Dialog.Panel className="flex flex-col relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-[31.25rem] sm:p-6">
        <UserList
          newRoomUsers={newRoomUsers}
          removeNewRoomUser={removeNewRoomUser}
        />
        <UserAutocompleteBox
          addNewRoomUser={addNewRoomUser}
          newRoomUsersDisplaySet={newRoomUsersDisplaySet}
        />
        <button
          type="button"
          disabled={!newRoomUsers.length}
          className="flex self-end mt-6 items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
          onClick={handleCreateNewRoom}
        >
          Create Room
        </button>
      </Dialog.Panel>
    </Transition.Child>
  );
};
