import { getAuthenticatedUser, getNewTextByUserid } from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage } from "next";

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: () => getAuthenticatedUser(),
    retry: false,
  });

  const { data: textData } = useQuery(
    ["text"],
    () =>
      getNewTextByUserid(authenticatedUserData?.id as number, {
        query: {
          specialCharactersGte: 5,
          specialCharactersLte: 10,
          numbersGte: 5,
          numbersLte: 10,
          punctuation: true,
          language: "en",
        },
      }),
    { enabled: !!authenticatedUserData?.id }
  );

  return <>{JSON.stringify(textData)}</>;
};

TrainPage.getInitialProps = async (ctx) => {
  const queryClient = new QueryClient();

  return {
    dehydratedState: dehydrate(queryClient),
  };
};

export default TrainPage;
