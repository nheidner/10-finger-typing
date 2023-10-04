import type { User } from "@/types";

export const Avatar = ({ user }: { user?: User }) => {
  if (!user) {
    return null;
  }

  return (
    <span className="inline-flex h-48 w-48 items-center justify-center rounded-full bg-gray-300">
      <span className="text-8xl font-medium leading-none text-white">
        {user.username?.[0].toUpperCase()}
      </span>
    </span>
  );
};
