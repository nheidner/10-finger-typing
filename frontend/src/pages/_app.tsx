import "@/styles/globals.css";
import {
  Hydrate,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import type { AppProps } from "next/app";
import { useState } from "react";
import { Header } from "../components/Header";
import { useRouter } from "next/router";
import { FetchError } from "@/utils/fetch";

const App = ({ Component, pageProps }: AppProps) => {
  const router = useRouter();

  const [queryClient] = useState(() => {
    const client = new QueryClient();

    client.setDefaultOptions({
      queries: {
        onError: (error: any) => {
          if (error instanceof FetchError && error.status === 401) {
            router.push("/login");
          }
        },
      },
      mutations: {
        onError: (error: any) => {
          if (error instanceof FetchError && error.status === 401) {
            router.push("/login");
          }
        },
      },
    });

    return client;
  });

  return (
    <QueryClientProvider client={queryClient}>
      <ReactQueryDevtools initialIsOpen={false} />
      <Hydrate state={pageProps.dehydratedState}>
        <main className="max-w-7xl mx-auto px-10 box-content">
          <Header />
          <Component {...pageProps} />
        </main>
      </Hydrate>
    </QueryClientProvider>
  );
};

export default App;
