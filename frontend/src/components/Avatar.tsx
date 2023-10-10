import type { User } from "@/types";
import classNames from "classnames";

export const Avatar = ({
  user,
  textClassName,
  containerClassName,
}: {
  textClassName: string;
  containerClassName: string;
  user?: Partial<User>;
}) => {
  if (!user) {
    return null;
  }

  const userDisplay = user?.username || user?.email;

  return (
    <span
      className={classNames(
        "inline-flex items-center justify-center rounded-full bg-gray-300",
        containerClassName
      )}
    >
      <span
        className={classNames(
          "font-medium leading-none text-white",
          textClassName
        )}
      >
        {userDisplay?.[0].toUpperCase()}
      </span>
    </span>
  );
};
