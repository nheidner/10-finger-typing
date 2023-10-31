import { debounce } from "@/utils/debounce";
import { getUsersByUsernamePartial } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { useRef, useState } from "react";

export const useDebouncedUserSearchByUsernamePartial = (
  debounceTime: number
) => {
  const [queryKey, setQueryKey] = useState("");
  const debouncedSetQueryKeyRef = useRef(
    debounce<void>(setQueryKey, debounceTime)
  );

  const handleUsernamePartialChange = (usernamePartial: string) => {
    debouncedSetQueryKeyRef.current(usernamePartial);
  };

  const { data, isFetching } = useQuery({
    queryKey: ["users", "username_contains", queryKey],
    queryFn: async () => getUsersByUsernamePartial(queryKey),
    retry: false,
    keepPreviousData: true,
  });

  return {
    users: data,
    isLoading: isFetching,
    debouncedFetchUsers: handleUsernamePartialChange,
  };
};
