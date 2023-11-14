import { useQuery } from "@tanstack/react-query";
import { FC, useEffect, useRef, useState } from "react";
import {
  getAuthenticatedUser,
  getRealtimeUserNotification,
} from "@/utils/queries";
import { UserNotification } from "@/types";
import Link from "next/link";

export const Notifications = () => {
  const [notifications, setNotifications] = useState<UserNotification[] | null>(
    null
  );

  const { data, isError } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: () => getAuthenticatedUser(),
    retry: false,
  });

  const userIsLoggedIn = !isError && !!data;

  const lastIdRef = useRef(Date.now().toString());
  const timeOutsRef = useRef<NodeJS.Timeout[]>([]);

  const { refetch } = useQuery(
    ["notification", lastIdRef.current],
    () => getRealtimeUserNotification(lastIdRef.current),
    {
      onSuccess: (notification) => {
        if (notification) {
          setNotifications((oldNotifications) =>
            oldNotifications
              ? oldNotifications.concat(notification)
              : [notification]
          );

          const closeNotificationTimeout = setTimeout(() => {
            closeNotification(notification.id);
          }, 10000);
          timeOutsRef.current.push(closeNotificationTimeout);

          lastIdRef.current = notification.id;
        }

        refetch();
      },
      enabled: userIsLoggedIn,
    }
  );

  useEffect(() => {
    const currentTimeouts = timeOutsRef.current;

    return () => {
      currentTimeouts.forEach((timeOut) => {
        clearTimeout(timeOut);
      });
    };
  }, []);

  const closeNotification = (notificationId: string) => {
    setNotifications((oldNotifications) =>
      oldNotifications
        ? oldNotifications.filter((notification) => {
            return notificationId !== notification.id;
          })
        : oldNotifications
    );
  };

  if (!notifications) {
    return null;
  }

  return (
    <div className="fixed flex flex-col gap-3 w-max max-w-full sm:max-w-[80%] md:max-w-[70%] lg:max-w-[50%] top-0 left-1/2 -translate-x-1/2 z-50 mt-4">
      {notifications.map((notification) => {
        let text = <></>;
        switch (notification.type) {
          case "room_invitation":
            text = (
              <>
                You were in invited by the user{" "}
                <Link href={`/${notification.payload.by}`}>
                  {notification.payload.by}
                </Link>{" "}
                to{" "}
                <Link href={`/rooms/${notification.payload.roomId}`}>
                  this room.
                </Link>
              </>
            );
            break;
        }

        return (
          <div
            id="toast-default"
            className="flex items-center p-4 text-gray-500 bg-white rounded-lg shadow dark:text-gray-400 dark:bg-gray-800"
            role="alert"
            key={notification.id}
          >
            <FireIcon />
            <div className="ms-3 text-sm font-normal">{text}</div>
            <XButton handleClose={() => closeNotification(notification.id)} />
          </div>
        );
      })}
    </div>
  );
};

const FireIcon = () => {
  return (
    <div className="inline-flex items-center justify-center flex-shrink-0 w-8 h-8 text-blue-500 bg-blue-100 rounded-lg dark:bg-blue-800 dark:text-blue-200">
      <svg
        className="w-4 h-4"
        aria-hidden="true"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 18 20"
      >
        <path
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M15.147 15.085a7.159 7.159 0 0 1-6.189 3.307A6.713 6.713 0 0 1 3.1 15.444c-2.679-4.513.287-8.737.888-9.548A4.373 4.373 0 0 0 5 1.608c1.287.953 6.445 3.218 5.537 10.5 1.5-1.122 2.706-3.01 2.853-6.14 1.433 1.049 3.993 5.395 1.757 9.117Z"
        />
      </svg>
      <span className="sr-only">Fire icon</span>
    </div>
  );
};

const XButton: FC<{ handleClose: () => void }> = ({ handleClose }) => {
  return (
    <button
      type="button"
      className="ms-auto ml-1.5 -mr-1.5 -my-1.5 bg-white text-gray-400 hover:text-gray-900 rounded-lg focus:ring-2 focus:ring-gray-300 p-1.5 hover:bg-gray-100 inline-flex items-center justify-center h-8 w-8 dark:text-gray-500 dark:hover:text-white dark:bg-gray-800 dark:hover:bg-gray-700"
      data-dismiss-target="#toast-default"
      aria-label="Close"
      onClick={handleClose}
    >
      <span className="sr-only">Close</span>
      <svg
        className="w-3 h-3"
        aria-hidden="true"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 14 14"
      >
        <path
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
        />
      </svg>
    </button>
  );
};
