import { getUserByUsername } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { Avatar } from "./Avatar";
import { ProfileData } from "./ProfileData";

export const Profile = ({ username }: { username: string }) => {
  const { data } = useQuery({
    queryKey: ["user", username],
    queryFn: () => getUserByUsername(username),
  });

  if (!data) {
    return null;
  }

  return (
    <section className="flex gap-[10%] border-b border-gray-100 mb-6 pb-8">
      <Avatar user={data} />
      <ProfileData user={data} />
    </section>
  );
};
