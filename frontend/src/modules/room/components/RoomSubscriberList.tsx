import { Avatar } from "@/components/Avatar";
import { RoomSubscriber, SubscriberGameStatus, User } from "@/types";
import { getAuthenticatedUser } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import classNames from "classnames";
import { FC, useMemo } from "react";

const statusToSortingWeightMapping: {
  [key in SubscriberGameStatus | "inactive"]: number;
} = {
  inactive: 0,
  unstarted: 1,
  started: 2,
  finished: 3,
};
const statusToColorMapping: {
  [key in SubscriberGameStatus | "inactive"]: string | undefined;
} = {
  inactive: undefined,
  unstarted: "#aaa",
  started: "#77f",
  finished: "#7f7",
};

export const RoomSubscriberList: FC<{ roomSubscribers: RoomSubscriber[] }> = ({
  roomSubscribers,
}) => {
  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: getAuthenticatedUser,
    retry: false,
  });

  const sortedRoomSubscribers = useMemo(() => {
    return roomSubscribers
      .filter((roomSubscriber) => {
        return roomSubscriber.userId != authenticatedUserData?.id;
      })
      ?.sort((roomSubscriberA, roomSubscriberB) => {
        const roomSubscriberAStatusA =
          roomSubscriberA.status === "inactive"
            ? "inactive"
            : roomSubscriberA.gameStatus;
        const roomSubscriberAStatusB =
          roomSubscriberB.status === "inactive"
            ? "inactive"
            : roomSubscriberB.gameStatus;

        return (
          statusToSortingWeightMapping[roomSubscriberAStatusB] -
          statusToSortingWeightMapping[roomSubscriberAStatusA]
        );
      });
  }, [roomSubscribers, authenticatedUserData?.id]);

  return (
    <ul className="flex gap-1 items-center">
      {sortedRoomSubscribers.map((roomSubscriber) => {
        const user: Partial<User> = {
          id: roomSubscriber.userId,
          username: roomSubscriber.username,
        };

        const status =
          roomSubscriber.status === "inactive"
            ? "inactive"
            : roomSubscriber.gameStatus;

        const color = statusToColorMapping[status];

        return (
          <li
            key={roomSubscriber.userId}
            style={{ borderColor: color }}
            className={classNames(
              "rounded-full list-none",
              color ? "border-2" : "p-[2px]"
            )}
          >
            <Avatar
              user={user}
              textClassName="text-1xl"
              containerClassName="w-10 h-10"
            />
          </li>
        );
      })}
    </ul>
  );
};
