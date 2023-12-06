import { KeyboardEvent, FC, FormEvent, useMemo } from "react";
import { User } from "@/types";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  NewRoomBodyParams,
  createRoomAndText,
  getAuthenticatedUser,
} from "@/utils/queries";
import { useRouter } from "next/router";
import { UserAutocompleteBox } from "./UserAutocompleteBox";
import { UserList } from "./UserList";
import { LoadingSpinner } from "@/components/LoadingSpinner";

const SubmitButton: FC<{ isDisabled: boolean; children: string }> = ({
  isDisabled,
  children,
}) => {
  return (
    <button
      type="submit"
      disabled={isDisabled}
      className="flex items-center rounded-md ml-2 bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
    >
      {children}
    </button>
  );
};

const sanitizeNewRoomUsers = (users: Partial<User>[]): NewRoomBodyParams => {
  const emails: string[] = [];
  const userIds: string[] = [];

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

  const { mutate: createNewRoomAndGameMutate, isLoading: createRoomIsLoading } =
    useMutation({
      mutationKey: ["create room and game"],
      mutationFn: createRoomAndText,
      onSuccess: (data) => {
        router.push({
          pathname: `rooms/${data.room.id}`,
        });
        removeNewRoomUsers();
        closeModal();
      },
    });

  const handleCreateNewRoomAndGame = (e: FormEvent) => {
    e.preventDefault();
    const params = sanitizeNewRoomUsers(newRoomUsers);

    createNewRoomAndGameMutate({
      newRoomBody: params,
      newGameBody: { textId },
    });
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter" && newRoomUsers.length) {
      handleCreateNewRoomAndGame(e);
    }
  };

  const submitButtonIsDisabled = !newRoomUsers.length;

  return (
    <form className="flex flex-col" onSubmit={handleCreateNewRoomAndGame}>
      <UserList
        newRoomUsers={newRoomUsers}
        removeNewRoomUser={removeNewRoomUser}
        isLoading={createRoomIsLoading}
      />
      <UserAutocompleteBox
        addNewRoomUser={addNewRoomUser}
        newRoomUsersDisplaySet={newRoomUsersDisplaySet}
        handleKeyDown={handleKeyDown}
      />
      <div className="self-end mt-6 flex items-center">
        <LoadingSpinner isLoading={createRoomIsLoading} />
        <SubmitButton isDisabled={submitButtonIsDisabled}>
          Create Room
        </SubmitButton>
      </div>
    </form>
  );
};
