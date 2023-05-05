import type { Book } from "@/types";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage } from "next";

const getBooks = async (id: string) =>
  fetch(`http://localhost:3000/api/users/${id}`).then(
    (res) => res.json() as Promise<{ data: Book[] }>
  );

const ProfilePage: NextPage<{
  userId: string;
  dehydratedState: DehydratedState;
}> = ({ userId }) => {
  const { data } = useQuery({
    queryKey: ["books"],
    queryFn: () => getBooks(userId),
  });

  return (
    <>
      <div>{JSON.stringify(data)}</div>
    </>
  );
};

ProfilePage.getInitialProps = async (ctx) => {
  const userId = ctx.req?.url?.split("/")[1] || "";
  const queryClient = new QueryClient();

  await queryClient.prefetchQuery(["books"], () => getBooks(userId));

  return {
    userId,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
