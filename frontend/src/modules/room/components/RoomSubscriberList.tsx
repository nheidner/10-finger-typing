import { Avatar } from "@/components/Avatar";
import {
  RoomSubscriber,
  SubscriberGameStatus,
  SubscriberStatus,
  User,
} from "@/types";
import { getAuthenticatedUser } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import classNames from "classnames";
import { FC, useMemo } from "react";
import { Flipper, Flipped } from "react-flip-toolkit";

const getSortingWeight = (
  subscriberStatus: SubscriberStatus,
  subscriberGameStatus: SubscriberGameStatus,
  cursorPosition?: number
): number => {
  if (subscriberStatus === "inactive") {
    return 0;
  }

  switch (subscriberGameStatus) {
    case "unstarted":
      return 1;
    case "started":
      return 2 + (cursorPosition || 0);
    case "finished":
      return Infinity;
  }
};

const statusToColorMapping: {
  [key in SubscriberGameStatus | "inactive"]: string | undefined;
} = {
  inactive: undefined,
  unstarted: "#aaa",
  started: "#77f",
  finished: "#7f7",
};

export const RoomSubscriberList: FC<{
  roomSubscribers: RoomSubscriber[];
  cursorPositions: Map<string, number>;
}> = ({ roomSubscribers, cursorPositions }) => {
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
        return (
          getSortingWeight(
            roomSubscriberB.status,
            roomSubscriberB.gameStatus,
            cursorPositions.get(roomSubscriberB.userId)
          ) -
          getSortingWeight(
            roomSubscriberA.status,
            roomSubscriberA.gameStatus,
            cursorPositions.get(roomSubscriberA.userId)
          )
        );
      });
  }, [roomSubscribers, authenticatedUserData?.id, cursorPositions]);

  return (
    <Flipper
      flipKey={sortedRoomSubscribers
        .map((roomSubscriber) => roomSubscriber.userId)
        .join(",")}
    >
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
            <Flipped key={roomSubscriber.userId} flipId={roomSubscriber.userId}>
              <li
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
            </Flipped>
          );
        })}
      </ul>
    </Flipper>
  );
};
