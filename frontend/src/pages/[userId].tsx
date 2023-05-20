import type { Book } from "@/types";
import { getApiUrl } from "@/utils/get_api_url";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage } from "next";

const getUsers = async (id: string) => {
  const apiUrl = getApiUrl();
  return fetch(`${apiUrl}/users/${id}`).then(
    (res) => res.json() as Promise<{ data: Book[] }>
  );
};

const ProfilePage: NextPage<{
  userId: string;
  dehydratedState: DehydratedState;
}> = ({ userId }) => {
  const { data } = useQuery({
    queryKey: ["users"],
    queryFn: () => getUsers(userId),
  });

  return (
    <>
      <div>{JSON.stringify(data)}</div>
    </>
  );
};

ProfilePage.getInitialProps = async (ctx) => {
  const { userId } = ctx.query as { userId: string };
  const queryClient = new QueryClient();

  await queryClient.prefetchQuery(["users"], () => getUsers(userId));

  return {
    userId,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
