import { getUserByUsername } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { Avatar } from "../../../components/Avatar";
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
      <Avatar
        user={data}
        textClassName="text-8xl"
        containerClassName="w-48 h-48"
      />

      <ProfileData user={data} />
    </section>
  );
};
